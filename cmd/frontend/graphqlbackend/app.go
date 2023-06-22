package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type LocalDirectoryArgs struct {
	Paths []string
}

type SetupNewAppRepositoriesForEmbeddingArgs struct {
	RepoNames []string
}

type EmbeddingSetupProgressArgs struct {
	RepoNames *[]string
}

type AppResolver interface {
	LocalDirectories(ctx context.Context, args *LocalDirectoryArgs) (LocalDirectoryResolver, error)
	LocalExternalServices(ctx context.Context) ([]LocalExternalServiceResolver, error)

	SetupNewAppRepositoriesForEmbedding(ctx context.Context, args SetupNewAppRepositoriesForEmbeddingArgs) (*EmptyResponse, error)
	EmbeddingsSetupProgress(ctx context.Context, args EmbeddingSetupProgressArgs) (EmbeddingsSetupProgressResolver, error)
}

type EmbeddingsSetupProgressResolver interface {
	OverallPercentComplete(ctx context.Context) (int32, error)
	CurrentRepository(ctx context.Context) *string
	CurrentRepositoryFilesProcessed(ctx context.Context) *int32
	CurrentRepositoryTotalFilesToProcess(ctx context.Context) *int32
}

type LocalDirectoryResolver interface {
	Paths() []string
	Repositories(ctx context.Context) ([]LocalRepositoryResolver, error)
}

type LocalRepositoryResolver interface {
	Name() string
	Path() string
}

type LocalExternalServiceResolver interface {
	ID() graphql.ID
	Path() string
	Autogenerated() bool
}
