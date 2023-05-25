package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	internalcontext "github.com/sourcegraph/sourcegraph/enterprise/internal/context"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func NewResolver(db database.DB, gitserverClient gitserver.Client, contextClient *internalcontext.ContextClient) graphqlbackend.ContextResolver {
	return &Resolver{
		db:              db,
		gitserverClient: gitserverClient,
		contextClient:   contextClient,
	}
}

type Resolver struct {
	db              database.DB
	gitserverClient gitserver.Client
	contextClient   *internalcontext.ContextClient
}

func (r *Resolver) GetContext(ctx context.Context, args graphqlbackend.GetContextArgs) ([]graphqlbackend.ContextResultResolver, error) {
	repoIDs, err := graphqlbackend.UnmarshalRepositoryIDs(args.Repos)
	if err != nil {
		return nil, err
	}

	repos, err := r.db.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	repoNameIDs := make([]types.RepoIDName, len(repoIDs))
	for i, repoID := range repoIDs {
		repoNameIDs[i] = types.RepoIDName{ID: repoID, Name: repos[repoID].Name}
	}

	fileChunks, err := r.contextClient.GetContext(ctx, internalcontext.GetContextArgs{
		Repos:            repoNameIDs,
		Query:            args.Query,
		CodeResultsCount: args.CodeResultsCount,
		TextResultsCount: args.TextResultsCount,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ContextResultResolver, len(fileChunks))
	for i, fileChunk := range fileChunks {
		resolvers[i], err = r.fileChunkToResolver(ctx, fileChunk)
		if err != nil {
			return nil, err
		}
	}

	return resolvers, nil
}

func (r *Resolver) fileChunkToResolver(ctx context.Context, chunk internalcontext.FileChunkContext) (graphqlbackend.ContextResultResolver, error) {
	repoResolver := graphqlbackend.NewRepositoryResolver(r.db, r.gitserverClient, &types.Repo{
		ID:   chunk.RepoID,
		Name: chunk.RepoName,
	})

	commitResolver := graphqlbackend.NewGitCommitResolver(r.db, r.gitserverClient, repoResolver, chunk.CommitID, nil)
	stat, err := r.gitserverClient.Stat(ctx, authz.DefaultSubRepoPermsChecker, chunk.RepoName, chunk.CommitID, chunk.Path)
	if err != nil {
		return nil, err
	}

	gitTreeEntryResolver := graphqlbackend.NewGitTreeEntryResolver(r.db, r.gitserverClient, graphqlbackend.GitTreeEntryResolverOpts{
		Commit: commitResolver,
		Stat:   stat,
	})
	return graphqlbackend.NewFileChunkContextResolver(gitTreeEntryResolver, chunk.StartLine, chunk.EndLine), nil
}
