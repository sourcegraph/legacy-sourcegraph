package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/stretchr/testify/assert"
)

const (
	// ReasonManualRepoSync and ReasonManualUserSync are copied from permssync
	// package to avoid import cycles.
	ReasonManualRepoSync = "REASON_MANUAL_REPO_SYNC"
	ReasonManualUserSync = "REASON_MANUAL_USER_SYNC"
)

func TestPermissionSyncJobs_CreateAndList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFakeClock(time.Now(), 0)

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Create(context.Background(), NewUser{Username: "horse"})
	assert.NoError(t, err)
	ctx := context.Background()

	store := PermissionSyncJobsWith(logger, db)

	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	assert.NoError(t, err)
	assert.Len(t, jobs, 0, "jobs returned even though database is empty")

	opts := PermissionSyncJobOpts{HighPriority: true, InvalidateCaches: true, Reason: ReasonManualRepoSync, TriggeredByUserID: user.ID}
	err = store.CreateRepoSyncJob(ctx, 99, opts)
	assert.NoError(t, err)

	nextSyncAt := clock.Now().Add(5 * time.Minute)
	opts = PermissionSyncJobOpts{HighPriority: false, InvalidateCaches: true, NextSyncAt: nextSyncAt, Reason: ReasonManualUserSync}
	err = store.CreateUserSyncJob(ctx, 77, opts)
	assert.NoError(t, err)

	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	assert.NoError(t, err)

	assert.Len(t, jobs, 2, "wrong number of jobs returned")

	wantJobs := []*PermissionSyncJob{
		{
			ID:                jobs[0].ID,
			State:             "queued",
			RepositoryID:      99,
			HighPriority:      true,
			InvalidateCaches:  true,
			Reason:            ReasonManualRepoSync,
			TriggeredByUserID: user.ID,
		},
		{
			ID:               jobs[1].ID,
			State:            "queued",
			UserID:           77,
			InvalidateCaches: true,
			ProcessAfter:     nextSyncAt,
			Reason:           ReasonManualUserSync,
		},
	}
	if diff := cmp.Diff(jobs, wantJobs, cmpopts.IgnoreFields(PermissionSyncJob{}, "QueuedAt")); diff != "" {
		t.Fatalf("jobs[0] has wrong attributes: %s", diff)
	}
	for i, j := range jobs {
		assert.NotZerof(t, j.QueuedAt, "job %d has no QueuedAt set", i)
	}

	listTests := []struct {
		name     string
		opts     ListPermissionSyncJobOpts
		wantJobs []*PermissionSyncJob
	}{
		{
			name:     "ID",
			opts:     ListPermissionSyncJobOpts{ID: jobs[0].ID},
			wantJobs: jobs[0:1],
		},
		{
			name:     "RepoID",
			opts:     ListPermissionSyncJobOpts{RepoID: jobs[0].RepositoryID},
			wantJobs: jobs[0:1],
		},
		{
			name:     "UserID",
			opts:     ListPermissionSyncJobOpts{UserID: jobs[1].UserID},
			wantJobs: jobs[1:],
		},
	}

	for _, tt := range listTests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := store.List(ctx, tt.opts)
			assert.NoError(t, err)
			if len(have) != len(tt.wantJobs) {
				t.Fatalf("wrong number of jobs returned. want=%d, have=%d", len(tt.wantJobs), len(have))
			}
			if diff := cmp.Diff(have, tt.wantJobs); diff != "" {
				t.Fatalf("unexpected jobs. diff: %s", diff)
			}
		})
	}
}

func TestPermissionSyncJobs_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFakeClock(time.Now(), 0)

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user1, err := db.Users().Create(context.Background(), NewUser{Username: "horse"})
	user2, err := db.Users().Create(context.Background(), NewUser{Username: "graph"})
	assert.NoError(t, err)
	ctx := context.Background()

	store := PermissionSyncJobsWith(logger, db)

	// 1) Insert low priority job without process_after for user1
	user1LowPrioJob := PermissionSyncJobOpts{Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	assert.NoError(t, err)
	allJobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	assert.NoError(t, err)
	// check that we have 1 job with userID=1
	assert.Len(t, allJobs, 1)
	assert.Equal(t, 1, allJobs[0].UserID)

	// 2) Insert low priority job without process_after for user2
	user2LowPrioJob := PermissionSyncJobOpts{Reason: ReasonManualUserSync, TriggeredByUserID: user2.ID}
	err = store.CreateUserSyncJob(ctx, 2, user2LowPrioJob)
	assert.NoError(t, err)
	allJobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	assert.NoError(t, err)
	// check that we have 2 jobs including job for userID=2. job ID should match user ID
	assert.Len(t, allJobs, 2)
	assert.Equal(t, allJobs[0].ID, allJobs[0].UserID)
	assert.Equal(t, allJobs[1].ID, allJobs[1].UserID)

	// 3) Another low priority job without process_after for user1 is dropped
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	assert.NoError(t, err)
	allJobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	assert.NoError(t, err)
	// check that we still have 2 jobs. Job ID should match user ID.
	assert.Len(t, allJobs, 2)
	assert.Equal(t, allJobs[0].ID, allJobs[0].UserID)
	assert.Equal(t, allJobs[1].ID, allJobs[1].UserID)

	// 4) Insert some low priority jobs with process_after for both users. All of them should be inserted.
	fiveMinutesLater := clock.Now().Add(5 * time.Minute)
	tenMinutesLater := clock.Now().Add(10 * time.Minute)
	user1LowPrioDelayedJob := PermissionSyncJobOpts{NextSyncAt: fiveMinutesLater, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	user2LowPrioDelayedJob := PermissionSyncJobOpts{NextSyncAt: tenMinutesLater, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioDelayedJob)
	assert.NoError(t, err)
	err = store.CreateUserSyncJob(ctx, 2, user2LowPrioDelayedJob)
	assert.NoError(t, err)
	allDelayedJobs, err := store.List(ctx, ListPermissionSyncJobOpts{NotNullProcessAfter: true})
	assert.NoError(t, err)
	// check that we have 2 delayed jobs in total
	assert.Len(t, allDelayedJobs, 2)
	// userID of the job should be (jobID - 2)
	assert.Equal(t, allDelayedJobs[0].UserID, allDelayedJobs[0].ID-2)
	assert.Equal(t, allDelayedJobs[1].UserID, allDelayedJobs[1].ID-2)

	// 5) Insert *high* priority job without process_after for user1. Check that low priority job is canceled
	user1HighPrioJob := PermissionSyncJobOpts{HighPriority: true, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	err = store.CreateUserSyncJob(ctx, 1, user1HighPrioJob)
	assert.NoError(t, err)
	allUser1Jobs, err := store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	assert.NoError(t, err)
	// check that we have 3 jobs for userID=1 in total (low prio (canceled), delayed, high prio)
	assert.Len(t, allUser1Jobs, 3)
	// check that low prio job (ID=1) is canceled and others are not
	for _, job := range allUser1Jobs {
		if job.ID == 1 {
			assert.True(t, job.Cancel)
		} else {
			assert.False(t, job.Cancel)
		}
	}

	// 6) Insert another low and high priority jobs without process_after for user1.
	// Check that all of them are dropped since we already have a high prio job.
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	assert.NoError(t, err)
	err = store.CreateUserSyncJob(ctx, 1, user1HighPrioJob)
	assert.NoError(t, err)
	allUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	assert.NoError(t, err)
	// check that we still have 3 jobs for userID=1 in total (low prio (canceled), delayed, high prio)
	assert.Len(t, allUser1Jobs, 3)

	// 7) Check that not "queued" jobs doesn't affect duplicates check: let's change high prio job to "processing"
	// and insert one low prio after that.
	result, err := db.ExecContext(ctx, "UPDATE permission_sync_jobs SET state='processing' WHERE id=5")
	assert.NoError(t, err)
	updatedRows, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), updatedRows)
	// Now we're good to insert new low prio job.
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	assert.NoError(t, err)
	allUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	assert.NoError(t, err)
	// check that we now have 4 jobs for userID=1 in total (low prio (canceled), delayed, high prio (processing), NEW low prio)
	assert.Len(t, allUser1Jobs, 4)
}
