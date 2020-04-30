package reader

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observability"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// ReaderMetrics encapsulates the Prometheus metrics of a Reader.
type ReaderMetrics struct {
	ReadMeta        *metrics.OperationMetrics
	ReadDocument    *metrics.OperationMetrics
	ReadResultChunk *metrics.OperationMetrics
	ReadDefinitions *metrics.OperationMetrics
	ReadReferences  *metrics.OperationMetrics
}

// MustRegister registers all metrics in ReaderMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (rm ReaderMetrics) MustRegister(r prometheus.Registerer) {
	for _, om := range []*metrics.OperationMetrics{
		rm.ReadMeta,
		rm.ReadDocument,
		rm.ReadResultChunk,
		rm.ReadDefinitions,
		rm.ReadReferences,
	} {
		om.MustRegister(prometheus.DefaultRegisterer)
	}
}

// NewReaderMetrics returns ReaderMetrics that need to be registered in a Prometheus registry.
func NewReaderMetrics(subsystem string) ReaderMetrics {
	return ReaderMetrics{
		ReadMeta:        metrics.NewOperationMetrics(subsystem, "reader", "read_meta"),
		ReadDocument:    metrics.NewOperationMetrics(subsystem, "reader", "read_document"),
		ReadResultChunk: metrics.NewOperationMetrics(subsystem, "reader", "read_result_chunk"),
		ReadDefinitions: metrics.NewOperationMetrics(subsystem, "reader", "read_definitions", metrics.WithCountHelp("The total number of definitions read")),
		ReadReferences:  metrics.NewOperationMetrics(subsystem, "reader", "read_references", metrics.WithCountHelp("The total number of references read")),
	}
}

// An ObservedReader wraps another Reader with error logging, Prometheus metrics, and tracing.
type ObservedReader struct {
	reader  Reader
	logger  logging.ErrorLogger
	metrics ReaderMetrics
	tracer  trace.Tracer
}

var _ Reader = &ObservedReader{}

// NewObservedReader wraps the given Reader with error logging, Prometheus metrics, and tracing.
func NewObservedReader(reader Reader, logger logging.ErrorLogger, metrics ReaderMetrics, tracer trace.Tracer) Reader {
	return &ObservedReader{
		reader:  reader,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// ReadMeta calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadMeta(ctx context.Context) (_ string, _ string, _ int, err error) {
	ctx, endTrace := r.prepTrace(ctx, &err, r.metrics.ReadMeta, "Reader.ReadMeta", "reader.read-meta")
	defer endTrace(1)

	return r.reader.ReadMeta(ctx)
}

// ReadDocument calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadDocument(ctx context.Context, path string) (_ types.DocumentData, _ bool, err error) {
	ctx, endTrace := r.prepTrace(ctx, &err, r.metrics.ReadDocument, "Reader.ReadDocument", "reader.read-document")
	defer endTrace(1)

	return r.reader.ReadDocument(ctx, path)
}

// ReadResultChunk calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadResultChunk(ctx context.Context, id int) (_ types.ResultChunkData, _ bool, err error) {
	ctx, endTrace := r.prepTrace(ctx, &err, r.metrics.ReadResultChunk, "Reader.ReadResultChunk", "reader.read-result-chunk")
	defer endTrace(1)

	return r.reader.ReadResultChunk(ctx, id)
}

// ReadDefinitions calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) (definitions []types.DefinitionReferenceRow, _ int, err error) {
	ctx, endTrace := r.prepTrace(ctx, &err, r.metrics.ReadDefinitions, "Reader.ReadDefinitions", "reader.read-definitions")
	defer func() { endTrace(float64(len(definitions))) }()

	return r.reader.ReadDefinitions(ctx, scheme, identifier, skip, take)
}

// ReadReferences calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) (references []types.DefinitionReferenceRow, _ int, err error) {
	ctx, endTrace := r.prepTrace(ctx, &err, r.metrics.ReadReferences, "Reader.ReadReferences", "reader.read-references")
	defer func() { endTrace(float64(len(references))) }()

	return r.reader.ReadReferences(ctx, scheme, identifier, skip, take)
}

func (r *ObservedReader) Close() error {
	return r.reader.Close()
}

func (r *ObservedReader) prepTrace(
	ctx context.Context,
	err *error,
	metrics *metrics.OperationMetrics,
	traceName string,
	logName string,
	preFields ...log.Field,
) (context.Context, observability.EndTraceFn) {
	return observability.PrepTrace(ctx, r.logger, metrics, r.tracer, err, traceName, logName, preFields...)
}
