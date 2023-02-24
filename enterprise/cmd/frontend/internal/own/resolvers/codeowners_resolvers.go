package resolvers

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// The Codeowners resolvers live under the parent Own resolver, but have their own file.
var (
	_ graphqlbackend.CodeownersIngestedFileResolver           = &codeownersIngestedFileResolver{}
	_ graphqlbackend.CodeownersIngestedFileConnectionResolver = &codeownersIngestedFileConnectionResolver{}
)

func (r *ownResolver) AddCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	proto, err := parseInputString(args.FileContents)
	if err != nil {
		return nil, err
	}

	codeownersFile := &types.CodeownersFile{
		RepoID:   api.RepoID(args.RepoID),
		Contents: args.FileContents,
		Proto:    proto,
	}
	if err := r.codeownersStore.CreateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not ingest codeowners file")
	}

	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
	}, nil
}

func (r *ownResolver) UpdateCodeownersFile(ctx context.Context, args *graphqlbackend.CodeownersFileArgs) (graphqlbackend.CodeownersIngestedFileResolver, error) {
	proto, err := parseInputString(args.FileContents)
	if err != nil {
		return nil, err
	}

	codeownersFile := &types.CodeownersFile{
		RepoID:   api.RepoID(args.RepoID),
		Contents: args.FileContents,
		Proto:    proto,
	}
	if err := r.codeownersStore.UpdateCodeownersFile(ctx, codeownersFile); err != nil {
		return nil, errors.Wrap(err, "could not update codeowners file")
	}

	return &codeownersIngestedFileResolver{
		codeownersFile: codeownersFile,
	}, nil
}

func parseInputString(fileContents string) (*codeownerspb.File, error) {
	fileReader := strings.NewReader(fileContents)
	proto, err := codeowners.Parse(fileReader)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse input")
	}
	return proto, nil
}

func (r *ownResolver) DeleteCodeownersFile(ctx context.Context, args *graphqlbackend.DeleteCodeownersFileArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.codeownersStore.DeleteCodeownersForRepo(ctx, api.RepoID(args.RepoID)); err != nil {
		return nil, errors.Wrapf(err, "could not delete codeowners file for repo %d", args.RepoID)
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *ownResolver) CodeownersIngestedFiles(ctx context.Context, args *graphqlbackend.CodeownersIngestedFilesArgs) (graphqlbackend.CodeownersIngestedFileConnectionResolver, error) {
	return nil, nil
}

type codeownersIngestedFileResolver struct {
	codeownersFile *types.CodeownersFile
}

func (c *codeownersIngestedFileResolver) Contents() string {
	return c.codeownersFile.Contents
}

func (c *codeownersIngestedFileResolver) RepoID() int32 {
	return int32(c.codeownersFile.RepoID)
}

func (c *codeownersIngestedFileResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: c.codeownersFile.CreatedAt}
}

func (c *codeownersIngestedFileResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: c.codeownersFile.UpdatedAt}
}

type codeownersIngestedFileConnectionResolver struct{}

func (c *codeownersIngestedFileConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CodeownersIngestedFileResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (c *codeownersIngestedFileConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (c *codeownersIngestedFileConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	//TODO implement me
	panic("implement me")
}
