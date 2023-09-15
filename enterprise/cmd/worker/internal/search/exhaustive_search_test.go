package search

import (
	"context"
	"io"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
)

func TestExhaustiveSearch(t *testing.T) {
	// This test exercises the full worker infra from the time a search job is
	// created until it is done.

	require := require.New(t)
	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := store.New(db, observation.TestContextTB(t))
	svc := service.New(observationCtx, store, mocks.NewMockStore())

	userID := insertRow(t, store.Store, "users", "username", "alice")
	insertRow(t, store.Store, "repo", "id", 1, "name", "repoa")
	insertRow(t, store.Store, "repo", "id", 2, "name", "repob")

	workerCtx, cancel1 := context.WithCancel(actor.WithInternalActor(context.Background()))
	defer cancel1()
	userCtx, cancel2 := context.WithCancel(actor.WithActor(context.Background(), actor.FromUser(userID)))
	defer cancel2()

	query := "1@rev1 1@rev2 2@rev3"

	// Create a job
	job, err := svc.CreateSearchJob(userCtx, query)
	require.NoError(err)

	// Do some assertations on the job before it runs
	{
		require.Equal(userID, job.InitiatorID)
		require.Equal(query, job.Query)
		require.Equal(types.JobStateQueued, job.State)
		require.NotZero(job.CreatedAt)
		require.NotZero(job.UpdatedAt)
		job2, err := svc.GetSearchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equal(job, job2)
	}

	// TODO these sort of tests need to live somewhere that makes more sense.
	// But for now we have a fully functioning setup here lets test List.
	{
		jobs, err := svc.ListSearchJobs(userCtx)
		require.NoError(err)
		require.Equal([]*types.ExhaustiveSearchJob{job}, jobs)
	}

	// Now that the job is created, we start up all the worker routines for
	// exhaustive search and wait until there are no more jobs left.
	searchJob := &searchJob{
		workerDB: db,
		config: config{
			WorkerInterval: 10 * time.Millisecond,
		},
	}

	// Each entry in bucket corresponds to one 1 uploaded csv file.
	mu := sync.Mutex{}
	var bucket []string

	mockStore := mocks.NewMockStore()
	mockStore.UploadFunc.SetDefaultHook(func(ctx context.Context, key string, r io.Reader) (int64, error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return 0, err
		}

		mu.Lock()
		bucket = append(bucket, string(b))
		mu.Unlock()

		return int64(len(b)), nil
	})

	routines, err := searchJob.newSearchJobRoutines(workerCtx, observationCtx, mockStore)
	require.NoError(err)
	for _, routine := range routines {
		go routine.Start()
		defer routine.Stop()
	}
	require.Eventually(func() bool {
		return !searchJob.hasWork(workerCtx)
	}, tTimeout(t, 10*time.Second), 10*time.Millisecond)

	// Assert that we ended up writing the expected results. This validates
	// that somehow the work happened (but doesn't dive into the guts of how
	// we co-ordinate our workers)
	{
		sort.Strings(bucket)
		require.Equal([]string{
			"repo,revspec,revision\n1,spec,rev1\n",
			"repo,revspec,revision\n1,spec,rev2\n",
			"repo,revspec,revision\n2,spec,rev3\n",
		}, bucket)
	}

	// Minor assertion that the job is regarded as finished.
	{
		job2, err := svc.GetSearchJob(userCtx, job.ID)
		require.NoError(err)
		// Only the WorkerJob fields should change. And in that case we will
		// only assert on State since the rest are non-deterministic.
		require.Equal(types.JobStateCompleted, job2.State)
		job2.WorkerJob = job.WorkerJob
		require.Equal(job, job2)
	}

	// Assert that cancellation affects the number of rows we expect. This is a bit
	// counterintuitive at this point because we have already completed the job.
	// However, cancellation affects the rows independently of the job state.
	{
		wantCount := 6
		gotCount, err := store.CancelSearchJob(userCtx, job.ID)
		require.NoError(err)
		require.Equal(wantCount, gotCount)
	}
}

// insertRow is a helper for inserting a row into a table. It assumes the
// table has an autogenerated column called id and it will return that value.
func insertRow(t testing.TB, store *basestore.Store, table string, keyValues ...any) int32 {
	var columns, values []*sqlf.Query
	for i, kv := range keyValues {
		if i%2 == 0 {
			columns = append(columns, sqlf.Sprintf(kv.(string)))
		} else {
			values = append(values, sqlf.Sprintf("%v", kv))
		}
	}
	q := sqlf.Sprintf(`INSERT INTO %s(%s) VALUES(%s) RETURNING id`, sqlf.Sprintf(table), sqlf.Join(columns, ", "), sqlf.Join(values, ", "))
	row := store.QueryRow(context.Background(), q)
	var id int32
	if err := row.Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// tTimeout returns the duration until t's deadline. If there is no deadline
// or the deadline is further away than max, then max is returned.
func tTimeout(t *testing.T, max time.Duration) time.Duration {
	deadline, ok := t.Deadline()
	if !ok {
		return max
	}
	timeout := time.Until(deadline)
	if max < timeout {
		return max
	}
	return timeout
}
