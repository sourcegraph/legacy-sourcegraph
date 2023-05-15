package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

func TestRepoCommits(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	repos := db.Repos()
	err := repos.Create(ctx, &types.Repo{ID: 1, Name: "foo"})
	require.NoError(t, err, "failed to insert repo")

	repoID := int32(1)

	commitSHA1 := "98d3ec26623660f17f6c298943f55aa339aa894a"
	commitSHA2 := "4b6152a804c4c177f5fe0dfd61e71cacb64d64dd"
	commitSHA3 := "e9c83398bbd4c4e9481fd20f100a7e49d68d7510"

	data := []types.PerforceChangelist{
		{
			CommitSHA:    api.CommitID(commitSHA1),
			ChangelistID: 123,
		},
		{
			CommitSHA:    api.CommitID(commitSHA2),
			ChangelistID: 124,
		},
		{
			CommitSHA:    api.CommitID(commitSHA3),
			ChangelistID: 125,
		},
	}

	s := RepoCommitsWith(logger, db)

	err = s.BatchInsertCommitSHAsWithPerforceChangelistID(ctx, api.RepoID(repoID), data)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("BatchInsertCommitSHAsWithPerforceChangelistID", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `SELECT repo_id, commit_sha, perforce_changelist_id FROM repo_commits ORDER by id`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		type repoCommitRow struct {
			RepoID       int32
			CommitSHA    string
			ChangelistID int64
		}

		want := []repoCommitRow{
			{
				RepoID:       int32(1),
				CommitSHA:    commitSHA1,
				ChangelistID: 123,
			},
			{
				RepoID:       int32(1),
				CommitSHA:    commitSHA2,
				ChangelistID: 124,
			},
			{
				RepoID:       int32(1),
				CommitSHA:    commitSHA3,
				ChangelistID: 125,
			},
		}

		got := []repoCommitRow{}

		for rows.Next() {
			var r repoCommitRow
			var hexCommitSHA []byte

			if err := rows.Scan(&r.RepoID, &hexCommitSHA, &r.ChangelistID); err != nil {
				t.Fatal(err)
			}

			r.CommitSHA = hex.EncodeToString(hexCommitSHA)
			got = append(got, r)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched rows, diff (-want, +got):\n %v\n", diff)
		}
	})

	t.Run("GetLatestForRepo", func(t *testing.T) {
		t.Run("existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLatestForRepo(ctx, api.RepoID(repoID))
			require.NoError(t, err, "unexpected error in GetLatestForRepo")
			require.NotNil(t, repoCommit, "repoCommit was not expected to be nil")
			require.Equal(
				t,
				&types.RepoCommit{
					ID:                   3,
					RepoID:               api.RepoID(repoID),
					CommitSHA:            dbutil.CommitBytea(commitSHA3),
					PerforceChangelistID: 125,
				},
				repoCommit,
				"repoCommit row is not as expected",
			)
		})

		t.Run("non existing repo", func(t *testing.T) {
			repoCommit, err := s.GetLatestForRepo(ctx, api.RepoID(2))
			require.Error(t, err)
			require.True(t, errors.Is(err, sql.ErrNoRows))
			require.Equal(t, &types.RepoCommit{}, repoCommit)
		})
	})
}
