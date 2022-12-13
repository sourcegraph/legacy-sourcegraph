// Package compression handles compressing the number of data points that need to be searched for a code insight series.
//
// The purpose is to reduce the extremely large number of search queries that need to run to backfill a historical insight.
//
// An index of commits is used to understand which time frames actually contain changes in a given repository.
// The commit index comes with metadata for each repository that understands the time at which the index was most recently updated.
// It is relevant to understand whether the index can be considered up to date for a repository or not, otherwise
// frames could be filtered out that simply are not yet indexed and otherwise should be queried.
//
// The commit indexer also has the concept of a horizon, that is to say the farthest date at which indices are stored. This horizon
// does not necessarily correspond to the last commit in the repository (the repo could be much older) so the compression must also
// understand this.
//
// At a high level, the algorithm is as follows:
//
// * Given a series of time frames [1....N]:
// * Always include 1 (to establish a baseline at the max horizon so that last observations may be carried)
// * For each remaining frame, check if it has commit metadata that is up to date, and check if it has no commits. If so, throw out the frame
// * Otherwise, keep the frame
package compression

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type NoopFilter struct {
}

type DataFrameFilter interface {
	Filter(ctx context.Context, sampleTimes []time.Time, name api.RepoName) BackfillPlan
}

type commitFetcher interface {
	RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error)
}

func NewGitserverFilter(db database.DB) DataFrameFilter {
	return &gitserverFilter{commitFetcher: discovery.NewGitCommitClient(db)}
}

type gitserverFilter struct {
	commitFetcher commitFetcher
}

// Filter will return a backfill plan that has filtered sample times for periods of time that do not change for a given repository.
func (g *gitserverFilter) Filter(ctx context.Context, sampleTimes []time.Time, name api.RepoName) BackfillPlan {
	var nodes []QueryExecution
	getCommit := func(to time.Time) (*gitdomain.Commit, bool, error) {
		commits, err := g.commitFetcher.RecentCommits(ctx, name, to)
		if err != nil {
			return nil, false, err
		} else if len(commits) == 0 {
			return nil, false, nil
		}

		return commits[0], true, nil
	}

	executions := make(map[api.CommitID]*QueryExecution)
	for _, sampleTime := range sampleTimes {
		commit, got, err := getCommit(sampleTime)
		if err != nil || !got {
			nodes = append(nodes, QueryExecution{RecordingTime: sampleTime})
		} else {
			qe, ok := executions[commit.ID]
			if ok {
				qe.SharedRecordings = append(qe.SharedRecordings, sampleTime)
			} else {
				executions[commit.ID] = &QueryExecution{
					Revision:      string(commit.ID),
					RecordingTime: sampleTime,
				}
			}
		}
	}

	for _, execution := range executions {
		nodes = append(nodes, *execution)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].RecordingTime.Before(nodes[j].RecordingTime)
	})

	return BackfillPlan{
		Executions:  nodes,
		RecordCount: len(nodes),
	}
}
func (n *NoopFilter) Filter(ctx context.Context, sampleTimes []time.Time, name api.RepoName) BackfillPlan {
	return uncompressedPlan(sampleTimes)
}

// uncompressedPlan returns a query plan that is completely uncompressed given an initial set of seed frames.
// This is primarily useful when there are scenarios in which compression cannot be used.
func uncompressedPlan(frames []time.Time) BackfillPlan {
	executions := make([]QueryExecution, 0, len(frames))
	for _, frame := range frames {
		executions = append(executions, QueryExecution{RecordingTime: frame})
	}

	return BackfillPlan{
		Executions:  executions,
		RecordCount: len(executions),
	}
}

// RecordCount returns the total count of data points that will be generated by this execution.
func (q *QueryExecution) RecordCount() int {
	return len(q.SharedRecordings) + 1
}

// ToRecording converts the query execution into a slice of recordable data points, each sharing the same value.
func (q *QueryExecution) ToRecording(seriesID string, repoName string, repoID api.RepoID, value float64) []store.RecordSeriesPointArgs {
	args := make([]store.RecordSeriesPointArgs, 0, q.RecordCount())
	base := store.RecordSeriesPointArgs{
		SeriesID: seriesID,
		Point: store.SeriesPoint{
			Time:  q.RecordingTime,
			Value: value,
		},
		RepoName:    &repoName,
		RepoID:      &repoID,
		PersistMode: store.RecordMode,
	}
	args = append(args, base)
	for _, sharedTime := range q.SharedRecordings {
		arg := base
		arg.Point.Time = sharedTime
		args = append(args, arg)
	}

	return args
}

// BackfillPlan is a rudimentary query plan. It provides a simple mechanism to store executable nodes
// to backfill an insight series.
type BackfillPlan struct {
	Executions  []QueryExecution
	RecordCount int
}

func (b BackfillPlan) String() string {
	var strs []string
	for i := range b.Executions {
		current := b.Executions[i]
		strs = append(strs, fmt.Sprintf("%v", current))
	}
	return fmt.Sprintf("[%v]", strings.Join(strs, ","))
}

// QueryExecution represents a node of an execution plan that should be queried against Sourcegraph.
// It can have dependent time points that will inherit the same value as the exemplar point
// once the query is executed and resolved.
type QueryExecution struct {
	Revision         string
	RecordingTime    time.Time
	SharedRecordings []time.Time
}
