package codenav

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CommitCache interface {
	AreCommitsResolvable(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
}

type commitCache struct {
	gitserverClient shared.GitserverClient
	mutex           sync.RWMutex
	cache           map[int]map[string]bool
}

func newCommitCache(client shared.GitserverClient) CommitCache {
	return &commitCache{
		gitserverClient: client,
		cache:           map[int]map[string]bool{},
	}
}

// AreCommitsResolvable determines if the given commits are resolvable for the given repositories.
// If we do not know the answer from a previous call to set or AreCommitsResolvable, we ask gitserver
// to resolve the remaining commits and store the results for subsequent calls. This method
// returns a slice of the same size as the input slice, true indicating that the commit at
// the symmetric index exists.
func (c *commitCache) AreCommitsResolvable(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error) {
	exists := make([]bool, len(commits))
	rcIndexMap := make([]int, 0, len(commits))
	rcs := make([]gitserver.RepositoryCommit, 0, len(commits))

	for i, rc := range commits {
		if e, ok := c.getInternal(rc.RepositoryID, rc.Commit); ok {
			exists[i] = e
		} else {
			rcIndexMap = append(rcIndexMap, i)
			rcs = append(rcs, gitserver.RepositoryCommit{
				RepositoryID: rc.RepositoryID,
				Commit:       rc.Commit,
			})
		}
	}

	// if there are no repository commits to fetch, we're done
	if len(rcs) == 0 {
		return exists, nil
	}

	// Perform heavy work outside of critical section
	e, err := c.gitserverClient.CommitsExist(ctx, rcs)
	if err != nil {
		return nil, errors.Wrap(err, "gitserverClient.CommitsExist")
	}
	if len(e) != len(rcs) {
		panic(strings.Join([]string{
			fmt.Sprintf("Expected slice returned from CommitsExist to have len %d, but has len %d.", len(rcs), len(e)),
			"If this panic occurred dcuring a test, your test is missing a mock definition for CommitsExist.",
			"If this is occurred during runtime, please file a bug.",
		}, " "))
	}

	for i, rc := range rcs {
		exists[rcIndexMap[i]] = e[i]
		c.setInternal(rc.RepositoryID, rc.Commit, e[i])
	}

	return exists, nil
}

func (c *commitCache) getInternal(repositoryID int, commit string) (bool, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if repositoryMap, ok := c.cache[repositoryID]; ok {
		if exists, ok := repositoryMap[commit]; ok {
			return exists, true
		}
	}

	return false, false
}

func (c *commitCache) setInternal(repositoryID int, commit string, exists bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.cache[repositoryID]; !ok {
		c.cache[repositoryID] = map[string]bool{}
	}

	c.cache[repositoryID][commit] = exists
}
