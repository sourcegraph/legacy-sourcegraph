package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// batchSpecWorkspaceExecutionJobStalledJobMaximumAge is the maximum allowable
// duration between updating the state of a job as "processing" and locking the
// record during processing. An unlocked row that is marked as processing
// likely indicates that the executor that dequeued the job has died. There
// should be a nearly-zero delay between these states during normal operation.
const batchSpecWorkspaceExecutionJobStalledJobMaximumAge = time.Second * 25

// batchSpecWorkspaceExecutionJobMaximumNumResets is the maximum number of
// times a job can be reset. If a job's failed attempts counter reaches this
// threshold, it will be moved into "failed" rather than "queued" on its next
// reset.
const batchSpecWorkspaceExecutionJobMaximumNumResets = 3

var batchSpecWorkspaceExecutionWorkerStoreOptions = dbworkerstore.Options{
	Name:              "batch_spec_workspace_execution_worker_store",
	TableName:         "batch_spec_workspace_execution_jobs",
	ColumnExpressions: batchSpecWorkspaceExecutionJobColumnsWithNullQueue.ToSqlf(),
	Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
		return scanFirstBatchSpecWorkspaceExecutionJob(rows, err)
	},
	OrderByExpression: sqlf.Sprintf("batch_spec_workspace_execution_jobs.place_in_global_queue"),
	StalledMaxAge:     batchSpecWorkspaceExecutionJobStalledJobMaximumAge,
	MaxNumResets:      batchSpecWorkspaceExecutionJobMaximumNumResets,
	// Explicitly disable retries.
	MaxNumRetries: 0,

	// This view ranks jobs from different users in a round-robin fashion
	// so that no single user can clog the queue.
	ViewName: "batch_spec_workspace_execution_queue batch_spec_workspace_execution_jobs",
}

type BatchSpecWorkspaceExecutionWorkerStore interface {
	dbworkerstore.Store

	FetchCanceled(ctx context.Context, executorName string) (canceledIDs []int, err error)
}

// NewBatchSpecWorkspaceExecutionWorkerStore creates a dbworker store that
// wraps the batch_spec_workspace_execution_jobs table.
func NewBatchSpecWorkspaceExecutionWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) BatchSpecWorkspaceExecutionWorkerStore {
	return &batchSpecWorkspaceExecutionWorkerStore{
		Store:              dbworkerstore.NewWithMetrics(handle, batchSpecWorkspaceExecutionWorkerStoreOptions, observationContext),
		observationContext: observationContext,
		accessTokenDeleterForTX: func(tx *Store) accessTokenHardDeleter {
			return tx.DatabaseDB().AccessTokens().HardDeleteByID
		},
	}
}

var _ dbworkerstore.Store = &batchSpecWorkspaceExecutionWorkerStore{}

// batchSpecWorkspaceExecutionWorkerStore is a thin wrapper around
// dbworkerstore.Store that allows us to extract information out of the
// ExecutionLogEntry field and persisting it to separate columns when marking a
// job as complete.
type batchSpecWorkspaceExecutionWorkerStore struct {
	dbworkerstore.Store

	accessTokenDeleterForTX func(tx *Store) accessTokenHardDeleter

	observationContext *observation.Context
}

func (s *batchSpecWorkspaceExecutionWorkerStore) FetchCanceled(ctx context.Context, executorName string) (canceledIDs []int, err error) {
	batchesStore := New(database.NewDBWith(s.Store), s.observationContext, nil)

	t := true
	cs, err := batchesStore.ListBatchSpecWorkspaceExecutionJobs(ctx, ListBatchSpecWorkspaceExecutionJobsOpts{
		Cancel:         &t,
		State:          btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		WorkerHostname: executorName,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(cs))
	for _, c := range cs {
		ids = append(ids, c.RecordID())
	}
	return ids, nil
}

type accessTokenHardDeleter func(context.Context, int64) error

// deleteAccessToken tries to delete the associated internal access
// token. If the token cannot be found it does *not* return an error.
func deleteAccessToken(ctx context.Context, deleteToken accessTokenHardDeleter, tokenID int64) error {
	err := deleteToken(ctx, tokenID)
	if err != nil && err != database.ErrAccessTokenNotFound {
		return err
	}
	return nil
}

type markFinal func(ctx context.Context, tx dbworkerstore.Store) (_ bool, err error)

func (s *batchSpecWorkspaceExecutionWorkerStore) markFinal(ctx context.Context, id int, fn markFinal) (ok bool, err error) {
	batchesStore := New(database.NewDBWith(s.Store), s.observationContext, nil)
	tx, err := batchesStore.Transact(ctx)
	if err != nil {
		return false, err
	}
	defer func() {
		// If we failed to mark the job as final, we fall back to the
		// non-wrapped functions so that the job does get marked as
		// final/errored if, e.g., deleting the access token failed.
		err = tx.Done(err)
		if err != nil {
			log15.Error("marking job as final failed, falling back to base method", "err", err)
			// Note: we don't use the transaction.
			ok, err = fn(ctx, s.Store)
		}
	}()

	job, err := tx.GetBatchSpecWorkspaceExecutionJob(ctx, GetBatchSpecWorkspaceExecutionJobOpts{ID: int64(id)})
	if err != nil {
		return false, err
	}

	err = deleteAccessToken(ctx, s.accessTokenDeleterForTX(tx), job.AccessTokenID)
	if err != nil {
		return false, err
	}

	events, err := logEventsFromLogEntries(job.ExecutionLogs)
	if err != nil {
		return false, err
	}

	executionResults, stepResults, err := extractCacheEntries(events)
	if err != nil {
		return false, err
	}

	workspace, err := tx.GetBatchSpecWorkspace(ctx, GetBatchSpecWorkspaceOpts{ID: job.BatchSpecWorkspaceID})
	if err != nil {
		return false, err
	}

	spec, err := tx.GetBatchSpec(ctx, GetBatchSpecOpts{ID: workspace.BatchSpecID})
	if err != nil {
		return false, err
	}

	for _, entry := range append(executionResults, stepResults...) {
		entry.UserID = spec.UserID
		if err := tx.CreateBatchSpecExecutionCacheEntry(ctx, entry); err != nil {
			return false, err
		}
	}

	return fn(ctx, s.Store.With(tx))
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkErrored(ctx context.Context, id int, failureMessage string, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	return s.markFinal(ctx, id, func(ctx context.Context, tx dbworkerstore.Store) (bool, error) {
		return tx.MarkErrored(ctx, id, failureMessage, options)
	})
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkFailed(ctx context.Context, id int, failureMessage string, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	return s.markFinal(ctx, id, func(ctx context.Context, tx dbworkerstore.Store) (bool, error) {
		return tx.MarkFailed(ctx, id, failureMessage, options)
	})
}

func (s *batchSpecWorkspaceExecutionWorkerStore) MarkComplete(ctx context.Context, id int, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	batchesStore := New(database.NewDBWith(s.Store), s.observationContext, nil)

	tx, err := batchesStore.Transact(ctx)
	if err != nil {
		return false, err
	}

	job, err := tx.GetBatchSpecWorkspaceExecutionJob(ctx, GetBatchSpecWorkspaceExecutionJobOpts{ID: int64(id)})
	if err != nil {
		return false, errors.Wrap(err, "loading batch spec workspace execution job")
	}

	workspace, err := tx.GetBatchSpecWorkspace(ctx, GetBatchSpecWorkspaceOpts{ID: job.BatchSpecWorkspaceID})
	if err != nil {
		return false, errors.Wrap(err, "loading batch spec workspace")
	}

	batchSpec, err := tx.GetBatchSpec(ctx, GetBatchSpecOpts{ID: workspace.BatchSpecID})
	if err != nil {
		return false, errors.Wrap(err, "loading batch spec")
	}

	// Impersonate as the user to ensure the repo is still accessible by them.
	ctx = actor.WithActor(ctx, actor.FromUser(batchSpec.UserID))

	repo, err := tx.Repos().Get(ctx, workspace.RepoID)
	if err != nil {
		return false, errors.Wrap(err, "loading repo")
	}

	events, err := logEventsFromLogEntries(job.ExecutionLogs)
	if err != nil {
		return false, err
	}

	rollbackAndMarkFailed := func(err error, fmtStr string, args ...any) (bool, error) {
		// Rollback transaction but ignore rollback errors. As long as nothing
		// is committed, we don't care.
		tx.Done(err)
		return s.Store.MarkFailed(ctx, id, fmt.Sprintf(fmtStr, args...), options)
	}

	executionResults, stepResults, err := extractCacheEntries(events)
	if err != nil {
		return rollbackAndMarkFailed(err, fmt.Sprintf("failed to extract cache entries: %s", err))
	}

	// This should never happen. If it does, guard against it.
	if len(executionResults) > 1 {
		return rollbackAndMarkFailed(errors.New("error"), "logs cannot contain more than 1 execution result")
	}

	for _, entry := range stepResults {
		// Store the cache entry.
		entry.UserID = batchSpec.UserID
		if err := tx.CreateBatchSpecExecutionCacheEntry(ctx, entry); err != nil {
			return rollbackAndMarkFailed(err, fmt.Sprintf("failed to save cache entry: %s", err))
		}
	}

	changesetSpecIDs := []int64{}
	for _, entry := range executionResults {
		// Store the cache entry.
		entry.UserID = batchSpec.UserID
		if err := tx.CreateBatchSpecExecutionCacheEntry(ctx, entry); err != nil {
			return rollbackAndMarkFailed(err, fmt.Sprintf("failed to save cache entry: %s", err))
		}

		// And now build changeset specs from it
		var executionResult execution.Result
		if err := json.Unmarshal([]byte(entry.Value), &executionResult); err != nil {
			return rollbackAndMarkFailed(err, fmt.Sprintf("failed to parse cache entry: %s", err))
		}

		rawSpecs, err := cache.ChangesetSpecsFromCache(
			batchSpec.Spec,
			batcheslib.Repository{
				ID:          string(relay.MarshalID("Repository", repo.ID)),
				Name:        string(repo.Name),
				BaseRef:     workspace.Branch,
				BaseRev:     workspace.Commit,
				FileMatches: workspace.FileMatches,
			},
			executionResult,
		)
		if err != nil {
			return rollbackAndMarkFailed(err, fmt.Sprintf("failed to build changeset specs from cache: %s", err))
		}

		var specs []*btypes.ChangesetSpec
		for _, rawSpec := range rawSpecs {
			changesetSpec, err := btypes.NewChangesetSpecFromSpec(rawSpec)
			if err != nil {
				return rollbackAndMarkFailed(err, fmt.Sprintf("failed to build db changeset specs: %s", err))
			}
			changesetSpec.BatchSpecID = batchSpec.ID
			changesetSpec.RepoID = repo.ID
			changesetSpec.UserID = batchSpec.UserID

			specs = append(specs, changesetSpec)
		}

		if len(specs) > 0 {
			if err := tx.CreateChangesetSpec(ctx, specs...); err != nil {
				return rollbackAndMarkFailed(err, fmt.Sprintf("failed to store changeset specs: %s", err))
			}
			for _, spec := range specs {
				changesetSpecIDs = append(changesetSpecIDs, spec.ID)
			}
		}
	}

	err = deleteAccessToken(ctx, s.accessTokenDeleterForTX(tx), job.AccessTokenID)
	if err != nil {
		return rollbackAndMarkFailed(err, fmt.Sprintf("failed to delete internal access token: %s", err))
	}

	if err = s.setChangesetSpecIDs(ctx, tx, job.BatchSpecWorkspaceID, changesetSpecIDs); err != nil {
		return rollbackAndMarkFailed(err, fmt.Sprintf("setChangesetSpecIDs: %s", err))
	}

	ok, err := s.Store.With(tx).MarkComplete(ctx, id, options)
	// When no record matched, we don't want to commit the rest. Make sure to properly fail the tx.
	if !ok && err == nil {
		return ok, tx.Done(errors.New("new"))
	}
	return ok, tx.Done(err)
}

func (s *batchSpecWorkspaceExecutionWorkerStore) setChangesetSpecIDs(ctx context.Context, tx *Store, batchSpecWorkspaceID int64, changesetSpecIDs []int64) error {
	m := make(map[int64]struct{}, len(changesetSpecIDs))
	for _, id := range changesetSpecIDs {
		m[id] = struct{}{}
	}
	marshaledIDs, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Set changeset_spec_ids on the batch_spec_workspace.
	res, err := tx.ExecResult(ctx, sqlf.Sprintf(setChangesetSpecIDsOnBatchSpecWorkspaceQueryFmtstr, marshaledIDs, batchSpecWorkspaceID))
	if err != nil {
		return err
	}
	c, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if c != 1 {
		return errors.New("incorrect number of changeset_specs updated")
	}
	return nil
}

const setChangesetSpecIDsOnBatchSpecWorkspaceQueryFmtstr = `
-- source: enterprise/internal/batches/store/worker_workspace_execution.go:setChangesetSpecIDs
UPDATE
	batch_spec_workspaces
SET
	changeset_spec_ids = %s
WHERE id = %s
`

func extractCacheEntries(events []*batcheslib.LogEvent) (executionResults, stepResults []*btypes.BatchSpecExecutionCacheEntry, err error) {
	for _, e := range events {
		switch m := e.Metadata.(type) {
		case *batcheslib.CacheResultMetadata:
			// TODO: This is stupid, because we unmarshal and then marshal again.
			value, err := json.Marshal(&m.Value)
			if err != nil {
				return nil, nil, err
			}

			executionResults = append(executionResults, &btypes.BatchSpecExecutionCacheEntry{
				Key:   m.Key,
				Value: string(value),
			})
		case *batcheslib.CacheAfterStepResultMetadata:
			// TODO: This is stupid, because we unmarshal and then marshal again.
			value, err := json.Marshal(&m.Value)
			if err != nil {
				return nil, nil, err
			}

			stepResults = append(stepResults, &btypes.BatchSpecExecutionCacheEntry{
				Key:   m.Key,
				Value: string(value),
			})
		}
	}

	return executionResults, stepResults, nil
}

var ErrNoSrcCLILogEntry = errors.New("no src-cli log entry found in execution logs")

func logEventsFromLogEntries(logs []workerutil.ExecutionLogEntry) ([]*batcheslib.LogEvent, error) {
	if len(logs) < 1 {
		return nil, errors.Newf("job has no execution logs")
	}

	var (
		entry workerutil.ExecutionLogEntry
		found bool
	)

	for _, e := range logs {
		if e.Key == "step.src.0" {
			entry = e
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNoSrcCLILogEntry
	}

	return btypes.ParseJSONLogsFromOutput(entry.Out), nil
}
