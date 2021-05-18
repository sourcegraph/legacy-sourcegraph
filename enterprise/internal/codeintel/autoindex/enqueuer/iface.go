package enqueuer

import (
	"context"
	"regexp"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	protocol "github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

type DBStore interface {
	basestore.ShareableStore

	Handle() *basestore.TransactableHandle
	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	DirtyRepositories(ctx context.Context) (map[int]int, error)
	IndexableRepositories(ctx context.Context, opts dbstore.IndexableRepositoryQueryOptions) ([]dbstore.IndexableRepository, error)
	UpdateIndexableRepository(ctx context.Context, indexableRepository dbstore.UpdateableIndexableRepository, now time.Time) error
	ResetIndexableRepositories(ctx context.Context, lastUpdatedBefore time.Time) error
	IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error)
	InsertIndex(ctx context.Context, index dbstore.Index) (int, error)
	RepoUsageStatistics(ctx context.Context) ([]dbstore.RepoUsageStatistics, error)
	GetRepositoriesWithIndexConfiguration(ctx context.Context) ([]int, error)
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (dbstore.IndexConfiguration, bool, error)
}

type DBStoreShim struct {
	*dbstore.Store
}

func (db *DBStoreShim) Transact(ctx context.Context) (DBStore, error) {
	store, err := db.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &DBStoreShim{store}, nil
}

var _ DBStore = &DBStoreShim{}

type RepoUpdaterClient interface {
	EnqueueRepoUpdate(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error)
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
}

type gitClient struct {
	client       GitserverClient
	repositoryID int
	commit       string
}

func newGitClient(client GitserverClient, repositoryID int, commit string) gitClient {
	return gitClient{
		client:       client,
		repositoryID: repositoryID,
		commit:       commit,
	}
}

func (s gitClient) ListFiles(ctx context.Context, pattern *regexp.Regexp) ([]string, error) {
	return s.client.ListFiles(ctx, s.repositoryID, s.commit, pattern)
}

func (s gitClient) FileExists(ctx context.Context, file string) (bool, error) {
	return s.client.FileExists(ctx, s.repositoryID, s.commit, file)
}

func (s gitClient) RawContents(ctx context.Context, file string) ([]byte, error) {
	return s.client.RawContents(ctx, s.repositoryID, s.commit, file)
}
