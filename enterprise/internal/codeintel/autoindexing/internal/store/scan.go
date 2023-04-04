package store

import (
	"database/sql"
	"sort"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

// scanIndexes scans a slice of indexes from the return value of `*Store.query`.
var scanIndexes = basestore.NewSliceScanner(scanIndex)

// scanFirstIndex scans a slice of indexes from the return value of `*Store.query` and returns the first.
var scanFirstIndex = basestore.NewFirstScanner(scanIndex)

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
	); err != nil {
		return index, err
	}

	index.ExecutionLogs = append(index.ExecutionLogs, executionLogs...)

	return index, nil
}

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

// scanFirstIndexConfiguration scans a slice of index configurations from the return value of `*Store.query`
// and returns the first.
var scanFirstIndexConfiguration = basestore.NewFirstScanner(scanIndexConfiguration)

func scanIndexConfiguration(s dbutil.Scanner) (indexConfiguration shared.IndexConfiguration, err error) {
	return indexConfiguration, s.Scan(
		&indexConfiguration.ID,
		&indexConfiguration.RepositoryID,
		&indexConfiguration.Data,
	)
}

// scanCounts scans pairs of id/counts from the return value of `*Store.query`.
func scanCounts(rows *sql.Rows, queryErr error) (_ map[int]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}

		visibilities[id] = count
	}

	return visibilities, nil
}

func scanCountsAndTotalCount(rows *sql.Rows, queryErr error) (totalCount int, _ map[int]int, err error) {
	if queryErr != nil {
		return 0, nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&totalCount, &id, &count); err != nil {
			return 0, nil, err
		}

		visibilities[id] = count
	}

	return totalCount, visibilities, nil
}

type sourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// scanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func scanSourcedCommits(rows *sql.Rows, queryErr error) (_ []sourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]sourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = sourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]sourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}

var ScanRepoRevs = basestore.NewSliceScanner(scanRepoRev)

func scanRepoRev(s dbutil.Scanner) (rr RepoRev, err error) {
	err = s.Scan(&rr.ID, &rr.RepositoryID, &rr.Rev)
	return rr, err
}
