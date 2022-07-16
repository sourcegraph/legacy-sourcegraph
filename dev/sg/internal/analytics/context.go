package analytics

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/otfields"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

// WithContext enables analytics in this context.
func WithContext(ctx context.Context, sgVersion string) (context.Context, error) {
	processor, err := newSpanToDiskProcessor()
	if err != nil {
		return ctx, errors.Wrap(err, "disk exporter")
	}

	provider := oteltracesdk.NewTracerProvider(
		oteltracesdk.WithResource(newResource(otfields.Resource{
			Name:       "sg",
			Namespace:  "dev",
			Version:    sgVersion,
			InstanceID: "", // TODO 'git config user.name' or 'user.email' maybe?
		})),
		oteltracesdk.WithSampler(oteltracesdk.AlwaysSample()),
		oteltracesdk.WithSpanProcessor(processor),
	)

	otel.SetTracerProvider(provider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		std.Out.WriteWarningf("opentelemetry: %s", err.Error())
	}))

	var rootSpan *Span
	ctx, rootSpan = StartSpan(ctx, "sg", trace.WithAttributes(
		attribute.Bool("root_span", true),
	))

	return context.WithValue(ctx, spansStoreKey{}, &spansStore{
		rootSpan: rootSpan.Span,
		provider: provider,
	}), nil
}

const (
	sgAnalyticsVersionResourceKey = "sg.analytics_version"
	// Increment to make breaking changes to spans and discard old spans
	sgAnalyticsVersion = "v1"
)

// newResource adapts sourcegraph/log.Resource into the OpenTelemetry package's Resource
// type.
func newResource(r log.Resource) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(r.Name),
		semconv.ServiceNamespaceKey.String(r.Namespace),
		semconv.ServiceInstanceIDKey.String(r.InstanceID),
		semconv.ServiceVersionKey.String(r.Version),
		attribute.String(sgAnalyticsVersionResourceKey, sgAnalyticsVersion))
}

func isValidVersion(spans *tracepb.ResourceSpans) bool {
	for _, attribute := range spans.GetResource().GetAttributes() {
		if attribute.GetKey() == sgAnalyticsVersionResourceKey {
			return attribute.Value.GetStringValue() == sgAnalyticsVersion
		}
	}
	return false
}
