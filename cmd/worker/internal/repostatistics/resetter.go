package repostatistics

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// resetter is a worker responsible for deleting all rows in
// `repo_statistics`/`gitserver_repos_statistics` and recreating their value.
//
// This is done because sometimes the statistics might get out of sync
// (possibly due to people modifying the `repo` table by hand)
type resetter struct{}

func NewResetter() job.Job {
	return &resetter{}
}

func (j *resetter) Description() string {
	return ""
}

func (j *resetter) Config() []env.Config {
	return nil
}

func (j *resetter) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			&resetterHandler{
				store:  db.RepoStatistics(),
				logger: observationCtx.Logger,
			},
			goroutine.WithName("repomgmt.statistics-resetter"),
			goroutine.WithDescription("resets repo statistics"),
			goroutine.WithInterval(30*time.Minute),
		),
	}, nil
}

type resetterHandler struct {
	store  database.RepoStatisticsStore
	logger log.Logger
}

var _ goroutine.Handler = &resetterHandler{}

func (h *resetterHandler) Handle(ctx context.Context) error {
	// We only run this handler once a week, Sunday morning between 2:00 and
	// 2:30 UTC, because it might run for 2-3 minutes.
	now := time.Now().UTC()
	isSunday := now.Weekday() == time.Sunday
	isBetween2And230 := now.Hour() == 2 && now.Minute() < 30

	if !isSunday || !isBetween2And230 {
		h.logger.Debug("Skipping deleting and recreating statistics; not Sunday between 2-2:30 UTC")
		return nil
	}

	// It's Sunday and between 2:00 and 2:30 AM UTC
	h.logger.Info("Deleting and recreating repo statistics")

	// We use a 5 minute timeout here to prevent the table lock from taking
	// down the instance.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return h.store.DeleteAndRecreateStatistics(ctx)
}
