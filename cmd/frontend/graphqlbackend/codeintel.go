package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// NewCodeIntelResolver will be set by enterprise
var NewCodeIntelResolver func() CodeIntelResolver

type CodeIntelResolver interface {
	LSIFUploadByID(ctx context.Context, id graphql.ID) (LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*EmptyResponse, error)
	LSIF(ctx context.Context, args *LSIFQueryArgs) (LSIFQueryResolver, error)
}

type LSIFUploadsQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	State           *string
	IsLatestForRepo *bool
	After           *string
}

type LSIFRepositoryUploadsQueryArgs struct {
	*LSIFUploadsQueryArgs
	RepositoryID graphql.ID
}

type LSIFUploadResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	InputCommit() string
	InputRoot() string
	State() string
	UploadedAt() DateTime
	StartedAt() *DateTime
	FinishedAt() *DateTime
	Failure() LSIFUploadFailureReasonResolver
	IsLatestForRepo() bool
}

type LSIFUploadFailureReasonResolver interface {
	Summary() string
	Stacktrace() string
}

type LSIFUploadConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFUploadResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LSIFQueryResolver interface {
	Commit(ctx context.Context) (*GitCommitResolver, error)
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
}

type LSIFQueryArgs struct {
	RepoID   api.RepoID
	RepoName api.RepoName
	Commit   GitObjectID
	Path     string
	UploadID int64
}

type LSIFQueryPositionArgs struct {
	Line      int32
	Character int32
}

type LSIFPagedQueryPositionArgs struct {
	LSIFQueryPositionArgs
	graphqlutil.ConnectionArgs
	After *string
}

type LocationConnectionResolver interface {
	Nodes(ctx context.Context) ([]LocationResolver, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type HoverResolver interface {
	Markdown() MarkdownResolver
	Range() RangeResolver
}

var codeIntelOnlyInEnterprise = errors.New("lsif uploads and queries are only available in enterprise")

func (r *schemaResolver) DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if EnterpriseResolvers.codeIntelResolver == nil {
		return nil, codeIntelOnlyInEnterprise
	}
	return EnterpriseResolvers.codeIntelResolver.DeleteLSIFUpload(ctx, args.ID)
}
