package codeownership

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type RulesKey struct {
	repoName api.RepoName
	commitID api.CommitID
}

type RulesCache struct {
	rules      map[RulesKey]*codeownerspb.File
	ownService backend.OwnService

	mu sync.RWMutex
}

func NewRulesCache(ownService backend.OwnService) RulesCache {
	return RulesCache{
		rules:      make(map[RulesKey]*codeownerspb.File),
		ownService: ownService,
	}
}

func (c *RulesCache) GetFromCacheOrFetch(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (*codeownerspb.File, error) {
	c.mu.RLock()
	key := RulesKey{repoName, commitID}
	if _, ok := c.rules[key]; ok {
		defer c.mu.RUnlock()
		return c.rules[key], nil
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	// Recheck condition.
	if _, ok := c.rules[key]; !ok {
		file, err := c.ownService.OwnersFile(ctx, repoName, commitID)
		if err != nil {
			emptyRuleset := &codeownerspb.File{}
			c.rules[key] = emptyRuleset
			return emptyRuleset, err
		}
		c.rules[key] = file
	}
	return c.rules[key], nil
}
