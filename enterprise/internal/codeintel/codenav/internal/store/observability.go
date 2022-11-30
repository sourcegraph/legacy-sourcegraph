package store

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// noop is a no-op operation to keep the newOperation scaffolding.
	noop *observation.Operation
}

var m = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		r,
		"codeintel_codenav_store",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	metrics, _ := m.Init(observationContext.Registerer)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		noop: op("noop"),
	}
}
