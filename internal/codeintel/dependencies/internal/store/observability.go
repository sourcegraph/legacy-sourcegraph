package store

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	deleteDependencyReposByID    *observation.Operation
	listDependencyRepos          *observation.Operation
	lockfileDependencies         *observation.Operation
	lockfileDependents           *observation.Operation
	preciseDependencies          *observation.Operation
	preciseDependents            *observation.Operation
	selectRepoRevisionsToResolve *observation.Operation
	updateResolvedRevisions      *observation.Operation
	upsertDependencyRepos        *observation.Operation
	upsertLockfileGraph          *observation.Operation
	listLockfileIndexes          *observation.Operation
	getLockfileIndex             *observation.Operation
}

var m = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		r,
		"codeintel_dependencies_store",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	metrics, _ := m.Init(observationContext.Registerer)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dependencies.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		deleteDependencyReposByID:    op("DeleteDependencyReposByID"),
		listDependencyRepos:          op("ListDependencyRepos"),
		lockfileDependencies:         op("LockfileDependencies"),
		lockfileDependents:           op("LockfileDependents"),
		preciseDependencies:          op("PreciseDependencies"),
		preciseDependents:            op("PreciseDependents"),
		selectRepoRevisionsToResolve: op("SelectRepoRevisionsToResolve"),
		updateResolvedRevisions:      op("UpdateResolvedRevisions"),
		upsertDependencyRepos:        op("UpsertDependencyRepos"),
		upsertLockfileGraph:          op("UpsertLockfileGraph"),
		listLockfileIndexes:          op("ListLockfileIndexes"),
		getLockfileIndex:             op("GetLockfileIndex"),
	}
}
