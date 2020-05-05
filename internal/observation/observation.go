// Package observation provides a unified way to wrap an operation with logging, tracing, and metrics.
//
// High-level ideas:
//
//     - Each service creates an observation Context that carries a root logger, tracer,
//       and a metrics registerer as its context.
//
//     - An observation Context can create an observation Operation which is configured
//       with log and trace names, a OperationMetrics value, and any log fields or metric
//       labels appropriate for the operation.
//
//     - An observation Operation can be prepared with its `With` method, which will prepare
//       a trace and some state to be reconciled later. This method returns a function that,
//       when deferred, will emit metrics, additional logs, and finalize the trace span.
//
// Sample usage:
//
//     observationContext := observation.NewContex(
//         log15.Root(),
//         &trace.Tracer{Tracer: opentracing.GlobalTracer()},
//         prometheus.DefaultRegisterer,
//     )
//
//     metrics := metrics.NewOperationMetrics(
//         "some_service",
//         "thing",
//         metrics.WithLabels("op"),
//     )
//
//     operation := observationContext.Operation(observation.Op{
//         LogName:      "Thing.SomeOperation",
//         TraceName:    "thing.some-operation",
//         MetricLabels: []string{"some_operation"},
//         Metrics:      metrics,
//     })
//
//     function SomeOperation(ctx context.Context) (err error) {
//         // logs and metrics may be available before or after the operation, so they
//         // can be supplied either at the start of the operation, or after in the
//         // defer of endObservation.
//
//         ctx, endObservation := operation.With(ctx, &err, observation.Args{ /* logs and metrics */ })
//         defer func() { endObservation(1, observation.Args{ /* additional logs and metrics */ }) }()
//
//         // ...
//     }
//
// Log fields and metric labels can be supplied at construction of an Operation, at invocation
// of an operation (the With function), or after the invocation completes but before the observation
// has terminated (the endObservation function). Log fields and metric labels are concatenated
// together in the order they are attached to an operation.
package observation

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Context holds the root objects that create loggers, trace spans, and registers metrics. Each
// service should create one context on start and use it to create Operation objects owned and
// used by the observed structs that perform the operation it represents.
type Context struct {
	logger     logging.ErrorLogger
	tracer     *trace.Tracer
	registerer prometheus.Registerer
}

// NewContext creates a new context. If the logger, tracer, or registerer passed here is nil,
// the operations created with this context will not have logging, tracing, or metrics enabled,
// respectively.
func NewContext(logger logging.ErrorLogger, tracer *trace.Tracer, registerer prometheus.Registerer) *Context {
	return &Context{
		logger:     logger,
		tracer:     tracer,
		registerer: registerer,
	}
}

// Op configures an Operation instance.
type Op struct {
	Metrics   *metrics.OperationMetrics
	TraceName string
	LogName   string
	// MetricLabels that apply for every invocation of this operation.
	MetricLabels []string
	// LogFields that apply for for every invocation of this operation.
	LogFields []log.Field
}

// Operation combines the state of the parent context to create a new operation. This value
// should be owned and used by the code that performs the operation it represents. This will
// immediately register any supplied metric with the context's metric registerer.
func (c *Context) Operation(args Op) *Operation {
	if c.registerer != nil && args.Metrics != nil {
		args.Metrics.MustRegister(c.registerer)
	}

	return &Operation{
		context:      c,
		metrics:      args.Metrics,
		traceName:    args.TraceName,
		logName:      args.LogName,
		metricLabels: args.MetricLabels,
		logFields:    args.LogFields,
	}
}

// Operation represents an interesting section of code that can be invoked.
type Operation struct {
	context      *Context
	metrics      *metrics.OperationMetrics
	traceName    string
	logName      string
	metricLabels []string
	logFields    []log.Field
}

// FinishFn is the shape of the function returned by With and should be invoked within
// a defer directly before the observed function returns.
type FinishFn func(count float64, args Args)

// Args configures the observation behavior of an invocation of an operation.
type Args struct {
	// MetricLabels that apply only to this invocation of the operation.
	MetricLabels []string
	// LogFields that apply only to this invocation of the operation.
	LogFields []log.Field
}

// With prepares the necessary timers, loggers, and metrics to observe the invocation of
// an operation.
func (op *Operation) With(ctx context.Context, err *error, args Args) (context.Context, FinishFn) {
	start := time.Now()
	tr, ctx := op.trace(ctx, args)

	return ctx, func(count float64, finishArgs Args) {
		elapsed := time.Since(start).Seconds()
		defaultFinishFields := []log.Field{log.Float64("count", count), log.Float64("elapsed", elapsed)}
		logFields := mergeLogFields(op.logFields, args.LogFields, defaultFinishFields, finishArgs.LogFields)
		metricLabels := mergeLabels(op.metricLabels, args.MetricLabels, finishArgs.MetricLabels)

		op.emitErrorLogs(err, logFields)
		op.emitMetrics(err, count, elapsed, metricLabels)
		op.finishTrace(err, tr, logFields)
	}
}

// trace creates a new Trace object and returns the wrapped context. If any log fields are
// attached to the operation or to the args to With, they are emitted immediately. This returns
// an unmodified context and a nil trace if no tracer was supplied on the observation context.
func (op *Operation) trace(ctx context.Context, args Args) (*trace.Trace, context.Context) {
	if op.context.tracer == nil {
		return nil, ctx
	}

	tr, ctx := op.context.tracer.New(ctx, op.traceName, "")
	tr.LogFields(mergeLogFields(op.logFields, args.LogFields)...)
	return tr, ctx
}

// emitErrorLogs will log as message if the operation has failed. This log contains the error
// as well as all of the log fields attached ot the operation, the args to With, and the args
// to the finish function. This does nothing if the no logger was supplied on the observation
// context.
func (op *Operation) emitErrorLogs(err *error, logFields []log.Field) {
	if op.context.logger == nil {
		return
	}

	var kvs []interface{}
	for _, field := range logFields {
		kvs = append(kvs, field.Key(), field.Value())
	}

	logging.Log(op.context.logger, op.logName, err, kvs...)
}

// emitMetrics will emit observe the duration, operation/result, and error counter metrics
// for this operation. This does nothing if no metric was supplied to the observation.
func (op *Operation) emitMetrics(err *error, count, elapsed float64, labels []string) {
	if op.metrics == nil {
		return
	}

	op.metrics.Observe(elapsed, count, err, labels...)
}

// finishTrace will set the error value, log additional fields supplied after the operation's
// execution, and finalize the trace span. This does nothing if no trace was constructed at
// the start of the operation.
func (op *Operation) finishTrace(err *error, tr *trace.Trace, logFields []log.Field) {
	if tr == nil {
		return
	}

	if err != nil {
		tr.SetError(*err)
	}

	tr.LogFields(logFields...)
	tr.Finish()
}

// mergeLabels flattens slices of slices of strings.
func mergeLabels(groups ...[]string) []string {
	size := 0
	for _, group := range groups {
		size += len(group)
	}

	labels := make([]string, 0, size)
	for _, group := range groups {
		labels = append(labels, group...)
	}

	return labels
}

// mergeLogFields flattens slices of slices of log fields.
func mergeLogFields(groups ...[]log.Field) []log.Field {
	size := 0
	for _, group := range groups {
		size += len(group)
	}

	logFields := make([]log.Field, 0, size)
	for _, group := range groups {
		logFields = append(logFields, group...)
	}

	return logFields
}
