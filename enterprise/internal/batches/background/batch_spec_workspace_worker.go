package background

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const batchSpecWorkspaceStalledJobMaximumAge = time.Second * 25
const batchSpecWorkspaceMaximumNumResets = 3

var batchSpecWorkspaceWorkerStoreOptions = dbworkerstore.Options{
	Name:              "batch_spec_workspace_worker_store",
	TableName:         "batch_spec_workspaces",
	ColumnExpressions: store.BatchSpecWorkspaceColums.ToSqlf(),
	Scan:              scanFirstBatchSpecWorkspaceRecord,
	OrderByExpression: sqlf.Sprintf("batch_spec_workspaces.created_at, batch_spec_workspaces.id"),
	StalledMaxAge:     batchSpecWorkspaceStalledJobMaximumAge,
	MaxNumResets:      batchSpecWorkspaceMaximumNumResets,
	// Explicitly disable retries.
	MaxNumRetries: 0,
}

type BatchSpecWorkspaceStore interface {
	dbworkerstore.Store
}

// NewExecutorStore creates a dbworker store that wraps the batch_spec_executions
// table.
func NewBatchSpecWorkspaceStore(handle *basestore.TransactableHandle, observationContext *observation.Context) BatchSpecWorkspaceStore {
	return &batchSpecWorkspaceStore{
		Store:              dbworkerstore.NewWithMetrics(handle, batchSpecWorkspaceWorkerStoreOptions, observationContext),
		observationContext: observationContext,
	}
}

var _ dbworkerstore.Store = &batchSpecWorkspaceStore{}

// batchSpecWorkspaceStore is a thin wrapper around dbworkerstore.Store that allows us to
// extract information out of the ExecutionLogEntry field and persisting it to
// separate columns when marking a job as complete.
type batchSpecWorkspaceStore struct {
	dbworkerstore.Store

	observationContext *observation.Context
}

// markCompleteQuery is taken from internal/workerutil/dbworker/store/store.go
//
// If that one changes we need to update this one here too.
const markBatchSpecWorkspaceCompleteQuery = `
UPDATE batch_spec_workspaces
SET state = 'completed', finished_at = clock_timestamp(), changeset_spec_ids = %s
WHERE id = %s AND state = 'processing' AND worker_hostname = %s
RETURNING id
`

func (s *batchSpecWorkspaceStore) MarkComplete(ctx context.Context, id int, options dbworkerstore.MarkFinalOptions) (_ bool, err error) {
	batchesStore := store.New(s.Store.Handle().DB(), s.observationContext, nil)

	changesetSpecIDs, err := loadAndExtractChangesetSpecIDs(ctx, batchesStore, int64(id))
	if err != nil {
		// If we couldn't extract the changeset IDs, we mark the job as failed
		return s.Store.MarkFailed(ctx, id, fmt.Sprintf("failed to extract changeset IDs ID: %s", err), options)
	}

	// TODO: This is copied from the batchesstore
	m := make(map[int64]struct{}, len(changesetSpecIDs))
	for _, id := range changesetSpecIDs {
		m[id] = struct{}{}
	}

	marshaledIDs, err := json.Marshal(m)
	if err != nil {
		return false, err
	}

	// TODO: Save batch_spec_id on changeset_specs
	_, ok, err := basestore.ScanFirstInt(batchesStore.Query(ctx, sqlf.Sprintf(markBatchSpecWorkspaceCompleteQuery, marshaledIDs, id, options.WorkerHostname)))
	return ok, err
}

func loadAndExtractChangesetSpecIDs(ctx context.Context, s *store.Store, id int64) ([]int64, error) {
	exec, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: id})
	if err != nil {
		return []int64{}, err
	}

	if len(exec.ExecutionLogs) < 1 {
		return []int64{}, errors.New("no execution logs")
	}

	randIDs, err := extractChangesetSpecRandIDs(exec.ExecutionLogs)
	if err != nil {
		return []int64{}, err
	}

	specs, _, err := s.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{LimitOpts: store.LimitOpts{Limit: 0}, RandIDs: randIDs})
	var ids []int64
	for _, spec := range specs {
		ids = append(ids, spec.ID)
	}

	return ids, nil
}

var ErrNoChangesetIDs = errors.New("no changeset ids found in execution logs")

func extractChangesetSpecRandIDs(logs []workerutil.ExecutionLogEntry) ([]string, error) {
	var (
		randIDs []string
		entry   workerutil.ExecutionLogEntry
		found   bool
	)

	for _, e := range logs {
		if e.Key == "step.src.0" {
			entry = e
			found = true
			break
		}
	}
	if !found {
		return randIDs, ErrNoChangesetIDs
	}

	for _, l := range strings.Split(entry.Out, "\n") {
		const outputLinePrefix = "stdout: "

		if !strings.HasPrefix(l, outputLinePrefix) {
			continue
		}

		jsonPart := l[len(outputLinePrefix):]

		var e changesetSpecsUploadedLogLine
		if err := json.Unmarshal([]byte(jsonPart), &e); err != nil {
			// If we can't unmarshal the line as JSON we skip it
			continue
		}

		if e.Operation == operationUploadingChangesetSpecs && e.Status == "SUCCESS" {
			rawIDs := e.Metadata.IDs
			if len(rawIDs) == 0 {
				return randIDs, ErrNoChangesetIDs
			}

			var randIDs []string
			for _, rawID := range rawIDs {
				var randID string
				if err := relay.UnmarshalSpec(graphql.ID(rawID), &randID); err != nil {
					return randIDs, errors.Wrap(err, "failed to unmarshal changeset spec rand id")
				}

				randIDs = append(randIDs, randID)
			}

			return randIDs, nil
		}
	}

	return randIDs, ErrNoBatchSpecRandID
}

type changesetSpecsUploadedLogLine struct {
	Operation string
	Timestamp time.Time
	Status    string
	Metadata  struct {
		IDs []string `json:"ids"`
	}
}

const operationUploadingChangesetSpecs = "UPLOADING_CHANGESET_SPECS"

func scanFirstBatchSpecWorkspaceRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpecWorkspace(rows, err)
}

// newBatchSpecWorkspaceResetter creates a dbworker.Resetter that re-enqueues
// lost batch_spec_workspace jobs for processing.
func newBatchSpecWorkspaceResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_spec_workspace_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.executionResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}
