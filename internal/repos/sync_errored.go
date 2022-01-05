package repos

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"golang.org/x/time/rate"

	"github.com/cockroachdb/errors"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

const syncInterval = 2 * time.Minute // TODO: decide on appropriate interval. Currently set low for ease of testing

func (s *Syncer) RunSyncReposWithLastErrorsWorker(ctx context.Context, rateLimiter *rate.Limiter) {
	for {
		log15.Info("running worker for SyncReposWithLastErrors", "time", time.Now())
		err := s.SyncReposWithLastErrors(ctx, rateLimiter)
		if err != nil {
			log15.Error("Error syncing repos w/ errors", "err", err)
		}

		// Wait and run task again
		time.Sleep(syncInterval)
	}
}

// SyncReposWithLastErrors iterates through all repos which have a non-empty last_error column in the gitserver_repos
// table, indicating there was an issue updating the repo, and syncs each of these repos. Repos which are no longer
// visible (i.e. deleted or made private) will be deleted from the DB. Note that this is only being run in Sourcegraph
// Dot com mode.
func (s *Syncer) SyncReposWithLastErrors(ctx context.Context, rateLimiter *rate.Limiter) error {
	return s.Store.GitserverReposStore.IterateWithNonemptyLastError(ctx, func(repo types.RepoGitserverStatus) error {
		err := rateLimiter.Wait(ctx)
		if err != nil {
			return errors.Errorf("error waiting for rate limiter: %s", err)
		}
		_, err = s.SyncRepo(ctx, repo.Name, false)
		if err != nil && !database.IsRepoNotFoundErr(err) {
			return err
		}
		return nil
	})
}
