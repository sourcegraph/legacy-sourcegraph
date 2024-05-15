package syntactic_indexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type syntacticIndexingSchedulerJob struct {
	Scheduler SyntacticJobScheduler
}

var _ job.Job = &syntacticIndexingSchedulerJob{}

var schedulerConfig *SchedulerConfig = &SchedulerConfig{}

func NewSyntacticindexingSchedulerJob() job.Job {
	return &syntacticIndexingSchedulerJob{}
}

func (j *syntacticIndexingSchedulerJob) Description() string {
	return ""
}

func (j *syntacticIndexingSchedulerJob) Config() []env.Config {
	return []env.Config{
		schedulerConfig,
	}
}

func (j *syntacticIndexingSchedulerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	rawDB, err := workerdb.InitRawDB(observationCtx)
	if err != nil {
		return nil, err
	}

	schedulerConfig.Load()

	scheduler, err := NewSyntacticJobScheduler(observationCtx, rawDB, schedulerConfig)

	if err != nil {
		return nil, err
	}

	j.Scheduler = scheduler

	return []goroutine.BackgroundRoutine{
		newSchedulerJob(
			observationCtx,
			scheduler,
		),
	}, nil

}

func newSchedulerJob(
	observationCtx *observation.Context,
	scheduler SyntacticJobScheduler,
) goroutine.BackgroundRoutine {

	m := new(metrics.SingletonREDMetrics)

	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_syntactic_indexing_background",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			config := conf.Get().ExperimentalFeatures

			if config != nil && config.CodeintelSyntacticIndexingEnabled {
				return scheduler.Schedule(observationCtx, ctx, time.Now())
			} else {
				observationCtx.Logger.Info("Syntactic indexing is disabled")
				return nil
			}
		}),
		goroutine.WithName("codeintel.syntactic-indexing-background-scheduler"),
		goroutine.WithDescription("schedule syntactic indexing jobs in the background"),
		goroutine.WithInterval(time.Second*5),
		goroutine.WithOperation(observationCtx.Operation(observation.Op{
			Name:              "codeintel.syntactic_indexing.HandleIndexSchedule",
			MetricLabelValues: []string{"HandleIndexSchedule"},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				return observation.EmitForDefault
			},
		})),
	)
}
