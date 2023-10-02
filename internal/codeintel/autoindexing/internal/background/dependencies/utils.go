package dependencies

import (
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// stalledIndexMaxAge is the maximum allowable duration between updating the state of an
// index as "processing" and locking the index row during processing. An unlocked row that
// is marked as processing likely indicates that the indexer that dequeued the index has
// died. There should be a nearly-zero delay between these states during normal operation.
const stalledIndexMaxAge = time.Second * 25

// indexMaxNumResets is the maximum number of times an index can be reset. If an index's
// failed attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const indexMaxNumResets = 3

var (
	AutoIndexingUseFifoAlgorithm = env.MustGetBool("CODEINTEL_AUTOINDEXING_DEQUEUE_FIFO_ALGORITHM", false, "Use the original FIFO dequeueing algorithm instead of the newer moving-window algorithm.")
	indexLookbackWindow          = env.Get("CODEINTEL_AUTOINDEXING_DEQUEUE_CANDIDATE_JOB_WINDOW", "1d", "The maximum age of index records prioritized for dequeue. Records older than this age will be visible for processing after the configured cooldown debuff.")
	repoDequeueCooldown          = env.Get("CODEINTEL_AUTOINDEXING_DEQUEUE_COOLDOWN_DEBUFF", "6h", "The minimum time since the last dequeue for a repository before records outside of the lookback window are made visible for processing. This allows older records of a repo to be processed, but with lower priority than records within the lookback window for other repositories.")
)

var IndexWorkerStoreOptions = func() dbworkerstore.Options[uploadsshared.Index] {
	if AutoIndexingUseFifoAlgorithm {
		return dbworkerstore.Options[uploadsshared.Index]{
			Name:              "codeintel_index",
			TableName:         "lsif_indexes",
			ViewName:          "lsif_indexes_with_repository_name u",
			ColumnExpressions: indexColumnsWithNullRank,
			Scan:              dbworkerstore.BuildWorkerScan(scanIndex),
			OrderByExpression: sqlf.Sprintf("(u.enqueuer_user_id > 0) DESC, u.queued_at DESC, u.id"),
			StalledMaxAge:     stalledIndexMaxAge,
			MaxNumResets:      indexMaxNumResets,
		}
	} else {
		return dbworkerstore.Options[uploadsshared.Index]{
			Name:              "codeintel_index",
			TableName:         "lsif_indexes",
			ViewName:          fmt.Sprintf("lsif_indexes_enqueue_candidates('%s'::interval, '%s'::interval) u", indexLookbackWindow, repoDequeueCooldown),
			ColumnExpressions: indexColumnsWithNullRank,
			Scan:              dbworkerstore.BuildWorkerScan(scanIndex),
			OrderByExpression: sqlf.Sprintf("TRUE"),
			StalledMaxAge:     stalledIndexMaxAge,
			MaxNumResets:      indexMaxNumResets,
		}
	}
}()

var indexColumnsWithNullRank = []*sqlf.Query{
	sqlf.Sprintf("u.id"),
	sqlf.Sprintf("u.commit"),
	sqlf.Sprintf("u.queued_at"),
	sqlf.Sprintf("u.state"),
	sqlf.Sprintf("u.failure_message"),
	sqlf.Sprintf("u.started_at"),
	sqlf.Sprintf("u.finished_at"),
	sqlf.Sprintf("u.process_after"),
	sqlf.Sprintf("u.num_resets"),
	sqlf.Sprintf("u.num_failures"),
	sqlf.Sprintf("u.repository_id"),
	sqlf.Sprintf("u.repository_name"),
	sqlf.Sprintf("u.docker_steps"),
	sqlf.Sprintf("u.root"),
	sqlf.Sprintf("u.indexer"),
	sqlf.Sprintf("u.indexer_args"),
	sqlf.Sprintf("u.outfile"),
	sqlf.Sprintf("u.execution_logs"),
	sqlf.Sprintf("NULL"),
	sqlf.Sprintf("u.local_steps"),
	sqlf.Sprintf("(SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id) AS associated_upload_id"),
	sqlf.Sprintf("u.should_reindex"),
	sqlf.Sprintf("u.requested_envvars"),
	sqlf.Sprintf("u.enqueuer_user_id"),
}

func scanIndex(s dbutil.Scanner) (index uploadsshared.Index, err error) {
	var executionLogs []executor.ExecutionLogEntry
	if err := s.Scan(
		&index.ID,
		&index.Commit,
		&index.QueuedAt,
		&index.State,
		&index.FailureMessage,
		&index.StartedAt,
		&index.FinishedAt,
		&index.ProcessAfter,
		&index.NumResets,
		&index.NumFailures,
		&index.RepositoryID,
		&index.RepositoryName,
		pq.Array(&index.DockerSteps),
		&index.Root,
		&index.Indexer,
		pq.Array(&index.IndexerArgs),
		&index.Outfile,
		pq.Array(&executionLogs),
		&index.Rank,
		pq.Array(&index.LocalSteps),
		&index.AssociatedUploadID,
		&index.ShouldReindex,
		pq.Array(&index.RequestedEnvVars),
		&index.EnqueuerUserID,
	); err != nil {
		return index, err
	}

	index.ExecutionLogs = append(index.ExecutionLogs, executionLogs...)

	return index, nil
}

// stalledDependencySyncingJobMaxAge is the maximum allowable duration between updating
// the state of a dependency indexing job as "processing" and locking the job row during
// processing. An unlocked row that is marked as processing likely indicates that the worker
// that dequeued the job has died. There should be a nearly-zero delay between these states
// during normal operation.
const stalledDependencySyncingJobMaxAge = time.Second * 25

// dependencySyncingJobMaxNumResets is the maximum number of times a dependency indexing
// job can be reset. If an job's failed attempts counter reaches this threshold, it will be
// moved into "errored" rather than "queued" on its next reset.
const dependencySyncingJobMaxNumResets = 3

var DependencySyncingJobWorkerStoreOptions = dbworkerstore.Options[dependencySyncingJob]{
	Name:              "codeintel_dependency_syncing",
	TableName:         "lsif_dependency_syncing_jobs",
	ColumnExpressions: dependencySyncingJobColumns,
	Scan:              dbworkerstore.BuildWorkerScan(scanDependencySyncingJob),
	OrderByExpression: sqlf.Sprintf("lsif_dependency_syncing_jobs.queued_at, lsif_dependency_syncing_jobs.upload_id"),
	StalledMaxAge:     stalledDependencySyncingJobMaxAge,
	MaxNumResets:      dependencySyncingJobMaxNumResets,
}

var dependencySyncingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("lsif_dependency_syncing_jobs.id"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.state"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.failure_message"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.started_at"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.finished_at"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.process_after"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.num_resets"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.num_failures"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.upload_id"),
}

func scanDependencySyncingJob(s dbutil.Scanner) (job dependencySyncingJob, err error) {
	return job, s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.UploadID,
	)
}

// stalledDependencyIndexingJobMaxAge is the maximum allowable duration between updating
// the state of a dependency indexing queueing job as "processing" and locking the job row during
// processing. An unlocked row that is marked as processing likely indicates that the worker
// that dequeued the job has died. There should be a nearly-zero delay between these states
// during normal operation.
const stalledDependencyIndexingJobMaxAge = time.Second * 25

// dependencyIndexingJobMaxNumResets is the maximum number of times a dependency indexing
// job can be reset. If an job's failed attempts counter reaches this threshold, it will be
// moved into "errored" rather than "queued" on its next reset.
const dependencyIndexingJobMaxNumResets = 3

var DependencyIndexingJobWorkerStoreOptions = dbworkerstore.Options[dependencyIndexingJob]{
	Name:              "codeintel_dependency_indexing",
	TableName:         "lsif_dependency_indexing_jobs",
	ColumnExpressions: dependencyIndexingJobColumns,
	Scan:              dbworkerstore.BuildWorkerScan(scanDependencyIndexingJob),
	OrderByExpression: sqlf.Sprintf("lsif_dependency_indexing_jobs.queued_at, lsif_dependency_indexing_jobs.upload_id"),
	StalledMaxAge:     stalledDependencyIndexingJobMaxAge,
	MaxNumResets:      dependencyIndexingJobMaxNumResets,
}

var dependencyIndexingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("lsif_dependency_indexing_jobs.id"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.state"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.failure_message"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.started_at"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.finished_at"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.process_after"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.num_resets"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.num_failures"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.upload_id"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.external_service_kind"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.external_service_sync"),
}

func scanDependencyIndexingJob(s dbutil.Scanner) (job dependencyIndexingJob, err error) {
	return job, s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.UploadID,
		&job.ExternalServiceKind,
		&job.ExternalServiceSync,
	)
}
