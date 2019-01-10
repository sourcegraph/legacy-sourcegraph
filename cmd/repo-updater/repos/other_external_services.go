package repos

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// OtherReposSyncer periodically synchronizes the configured repos in "OTHER" external service
// connections with the stored repos in Sourcegraph.
type OtherReposSyncer struct {
	// InternalAPI client used to fetch all external servicess and upsert repos.
	api InternalAPI
	// RWMutex synchronizing access to repos below
	mu sync.RWMutex
	// Latest synced repos cache used by GetRepoInfoByName.
	repos map[string]*protocol.RepoInfo
	// Channel passed in NewOtherReposSyncer where synced repos are sent to after being cached.
	synced chan<- *protocol.RepoInfo
}

// InternalAPI captures the internal API methods needed for syncing external services' repos.
type InternalAPI interface {
	ExternalServicesList(context.Context, api.ExternalServicesListRequest) ([]*api.ExternalService, error)
	ReposCreateIfNotExists(context.Context, api.RepoCreateOrUpdateRequest) (*api.Repo, error)
	ReposUpdateMetadata(ctx context.Context, repo api.RepoName, description string, fork, archived bool) error
}

// NewOtherReposSyncer returns a new OtherReposSyncer. Synced repos will be sent on the given channel.
func NewOtherReposSyncer(api InternalAPI, synced chan<- *protocol.RepoInfo) *OtherReposSyncer {
	return &OtherReposSyncer{
		api:    api,
		repos:  map[string]*protocol.RepoInfo{},
		synced: synced,
	}
}

// GetRepoInfoByName returns repo info of the repository with the given name.
func (s *OtherReposSyncer) GetRepoInfoByName(ctx context.Context, name string) *protocol.RepoInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repos[name]
}

// Run periodically synchronizes the configured repos in "OTHER" external service
// connections with the stored repos in Sourcegraph. Termination is done through the passed context.
func (s *OtherReposSyncer) Run(ctx context.Context, interval time.Duration) error {
	ticks := time.NewTicker(interval)
	defer ticks.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticks.C:
			log15.Debug("syncing all OTHER external services")

			results, err := s.syncAll(ctx)
			if err != nil {
				log15.Error("error syncing other external services repos", "error", err)
			}

			for _, err := range results.Errors() {
				log15.Error("sync error", err.Error())
			}
		}
	}
}

// syncAll syncrhonizes all "OTHER" external services.
func (s *OtherReposSyncer) syncAll(ctx context.Context) (SyncResults, error) {
	svcs, err := s.api.ExternalServicesList(ctx, api.ExternalServicesListRequest{Kind: "OTHER"})
	if err != nil {
		return nil, err
	}
	return s.SyncMany(ctx, svcs...), nil
}

// SyncResults is a helper type for lists of SyncResults.
type SyncResults []*SyncResult

// Errors returns all Errors in the list of SyncResults.
func (rs SyncResults) Errors() (errs []*SyncError) {
	for _, res := range rs {
		errs = append(errs, res.Errors...)
	}
	return errs
}

// SyncErrors is a helper type for lists of SyncErrors.
type SyncErrors []*SyncError

// Error implements the error interface.
func (errs SyncErrors) Error() string {
	var sb strings.Builder
	for _, err := range errs {
		sb.WriteString(err.Error() + "; ")
	}
	return sb.String()
}

// SyncError is an error type containing information about a failed sync of an external service
// of kind "OTHER".
type SyncError struct {
	// External service that had an error synchronizing.
	Service *api.ExternalService
	// Repo that failed synchronizing. This may be nil if the synchronization
	// process failed before attempting to sync each repo of defined by the external
	// service config.
	Repo *protocol.RepoInfo
	// The actual error.
	Err string
}

// Error implements the error interface.
func (e SyncError) Error() string {
	if e.Repo == nil {
		return fmt.Sprintf("external-service=%q: %s", e.Service.DisplayName, e.Err)
	}
	return fmt.Sprintf("external-service=%q repo=%q: %s", e.Service.DisplayName, e.Repo.Name, e.Err)
}

// SyncResult is returned by Sync to indicate which external services and their
// repos synced successfully and which didn't.
type SyncResult struct {
	// The external service of kind "OTHER" that had its repos synced.
	Service *api.ExternalService
	// Repos that succeeded to be synced.
	Synced []*protocol.RepoInfo
	// Repos that failed to be synced.
	Errors SyncErrors
}

// SyncMany synchonizes the repos defined by all the given external services of kind "OTHER".
// It return a SyncResults containing which repos were synced and which failed to.
func (s *OtherReposSyncer) SyncMany(ctx context.Context, svcs ...*api.ExternalService) SyncResults {
	if len(svcs) == 0 {
		return nil
	}

	ch := make(chan *SyncResult, len(svcs))
	for _, svc := range svcs {
		go func(svc *api.ExternalService) {
			ch <- s.Sync(ctx, svc)
		}(svc)
	}

	results := make([]*SyncResult, 0, len(svcs))
	for i := 0; i < cap(ch); i++ {
		res := <-ch
		results = append(results, res)
	}

	return results
}

// Sync synchronizes the repositories of a single external service of kind "OTHER"
func (s *OtherReposSyncer) Sync(ctx context.Context, svc *api.ExternalService) *SyncResult {
	cloneURLs, err := otherExternalServiceCloneURLs(svc)
	if err != nil {
		return &SyncResult{
			Service: svc,
			Errors:  []*SyncError{{Service: svc, Err: err.Error()}},
		}
	}

	repos := make([]*protocol.RepoInfo, 0, len(cloneURLs))
	for _, u := range cloneURLs {
		repoURL := u.String()
		repoName := otherRepoName(u)
		u.Path, u.RawQuery = "", ""
		serviceID := u.String()

		repos = append(repos, &protocol.RepoInfo{
			Name: repoName,
			VCS:  protocol.VCSInfo{URL: repoURL},
			ExternalRepo: &api.ExternalRepoSpec{
				ID:          string(repoName),
				ServiceType: "other",
				ServiceID:   serviceID,
			},
		})
	}

	res := s.store(ctx, svc, repos...)
	if len(res.Synced) > 0 {
		otherExternalServicesUpdateTime.WithLabelValues(svc.DisplayName).Set(float64(time.Now().Unix()))
	}

	return res
}

// store upserts the given repos through the FrontendAPI, returning which succeeded
// and which failed to be processed.
func (s *OtherReposSyncer) store(ctx context.Context, svc *api.ExternalService, repos ...*protocol.RepoInfo) *SyncResult {
	if len(repos) == 0 {
		return &SyncResult{Service: svc}
	}

	type storeOp struct {
		repo *protocol.RepoInfo
		err  error
	}

	ch := make(chan storeOp, len(repos))
	for _, repo := range repos {
		go func(op storeOp) {
			op.err = s.upsert(ctx, op.repo)
			ch <- op
		}(storeOp{repo: repo})
	}

	res := SyncResult{Service: svc}
	for i := 0; i < cap(ch); i++ {
		if op := <-ch; op.err != nil {
			res.Errors = append(res.Errors, &SyncError{
				Service: svc,
				Repo:    op.repo,
				Err:     op.err.Error(),
			})
		} else {
			res.Synced = append(res.Synced, op.repo)
			s.cache(op.repo)
		}
	}
	return &res
}

func (s *OtherReposSyncer) cache(repo *protocol.RepoInfo) {
	s.mu.Lock()
	s.repos[string(repo.Name)] = repo
	s.mu.Unlock()

	if s.synced != nil {
		s.synced <- repo
	}
}

func (s *OtherReposSyncer) upsert(ctx context.Context, repo *protocol.RepoInfo) error {
	_, err := s.api.ReposCreateIfNotExists(ctx, api.RepoCreateOrUpdateRequest{
		RepoName:     repo.Name,
		Enabled:      true,
		Fork:         repo.Fork,
		Archived:     repo.Archived,
		Description:  repo.Description,
		ExternalRepo: repo.ExternalRepo,
	})

	if err != nil {
		return err
	}

	return s.api.ReposUpdateMetadata(
		ctx,
		repo.Name,
		repo.Description,
		repo.Fork,
		repo.Archived,
	)
}

var otherRepoNameReplacer = strings.NewReplacer(":", "-", "@", "-", "//", "")

func otherRepoName(cloneURL *url.URL) api.RepoName {
	u := *cloneURL
	u.User = nil
	u.Scheme = ""
	u.RawQuery = ""
	u.Fragment = ""
	return api.RepoName(otherRepoNameReplacer.Replace(u.String()))
}

// otherExternalServiceCloneURLs returns all cloneURLs of the given "OTHER" external service.
func otherExternalServiceCloneURLs(s *api.ExternalService) ([]*url.URL, error) {
	var c schema.OtherExternalServiceConnection
	if err := jsonc.Unmarshal(s.Config, &c); err != nil {
		return nil, err
	}

	if len(c.Repos) == 0 {
		return nil, nil
	}

	parseRepo := url.Parse
	if c.Url != "" {
		baseURL, err := url.Parse(c.Url)
		if err != nil {
			return nil, err
		}
		parseRepo = baseURL.Parse
	}

	cloneURLs := make([]*url.URL, 0, len(c.Repos))
	for _, repo := range c.Repos {
		cloneURL, err := parseRepo(repo)
		if err != nil {
			log15.Error("skipping invalid repo clone URL", "repo", repo, "url", c.Url, "error", err)
			continue
		}
		cloneURLs = append(cloneURLs, cloneURL)
	}

	return cloneURLs, nil
}
