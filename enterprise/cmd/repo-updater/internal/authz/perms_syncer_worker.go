package authz

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func MakePermsSyncerWorker(ctx context.Context, syncer *PermsSyncer) *permsSyncerWorker {
	syncGroups := map[requestType]group.ContextGroup{
		requestTypeUser: group.New().WithContext(ctx).WithMaxConcurrency(syncUsersMaxConcurrency()),
		requestTypeRepo: group.New().WithContext(ctx).WithMaxConcurrency(1),
	}

	return &permsSyncerWorker{syncer: syncer, syncGroups: syncGroups}
}

type permsSyncerWorker struct {
	syncer     *PermsSyncer
	syncGroups map[requestType]group.ContextGroup
}

func (h *permsSyncerWorker) Handle(ctx context.Context, logger log.Logger, record *database.PermissionSyncJob) error {
	fmt.Printf("handling record: %+v\n", record)
	prio := priorityLow
	if record.HighPriority {
		prio = priorityHigh
	}

	reqType := requestTypeRepo
	if record.UserID != 0 {
		reqType = requestTypeUser
	}
	reqID := int32(record.RepositoryID)
	if record.UserID != 0 {
		reqID = int32(record.UserID)
	}

	// actually to the perm syncing
	h.syncer.syncPerms(ctx, h.syncGroups, &syncRequest{requestMeta: &requestMeta{
		Priority: prio,
		Type:     reqType,
		ID:       reqID,
		Options: authz.FetchPermsOptions{
			InvalidateCaches: record.InvalidateCaches,
		},
		NextSyncAt: time.Time{},
		NoPerms:    false,
	}})
	return nil
}

func MakeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*database.PermissionSyncJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*database.PermissionSyncJob]{
		Name:              "permission_sync_job_worker_store",
		TableName:         "permission_sync_jobs",
		ColumnExpressions: database.PermissionSyncJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(database.ScanPermissionSyncJob),
		OrderByExpression: sqlf.Sprintf("permission_sync_jobs.repository_id, permission_sync_jobs.user_id, permission_sync_jobs.high_priority"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 30,
	})
}

func MakeWorker(ctx context.Context, workerStore dbworkerstore.Store[*database.PermissionSyncJob], permsSyncer *PermsSyncer) *workerutil.Worker[*database.PermissionSyncJob] {
	handler := MakePermsSyncerWorker(ctx, permsSyncer)

	return dbworker.NewWorker[*database.PermissionSyncJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "permission_sync_job_worker",
		Interval:          time.Second, // Poll for a job once per second
		NumHandlers:       1,           // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
	})
}

func MakeResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*database.PermissionSyncJob]) *dbworker.Resetter[*database.PermissionSyncJob] {
	return dbworker.NewResetter(observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "permission_sync_job_worker_resetter",
		Interval: time.Second * 30, // Check for orphaned jobs every 30 seconds
		Metrics:  makeResetterMetrics(observationCtx, "permission_sync_job_worker"),
	})
}

// TODO: this function is copy pasted and should be made reusable
func makeResetterMetrics(observationCtx *observation.Context, workerName string) dbworker.ResetterMetrics {
	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_reset_failures_total", workerName),
		Help: "The number of reset failures.",
	})
	observationCtx.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_resets_total", workerName),
		Help: "The number of records reset.",
	})
	observationCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_reset_errors_total", workerName),
		Help: "The number of errors that occur when resetting records.",
	})
	observationCtx.Registerer.MustRegister(errors)

	return dbworker.ResetterMetrics{
		RecordResets:        resets,
		RecordResetFailures: resetFailures,
		Errors:              errors,
	}
}
