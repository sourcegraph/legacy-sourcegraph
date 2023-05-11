package resolvers

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingSearchResolver(t *testing.T) {
	logger := logtest.Scoped(t)

	mockDB := database.NewMockDB()
	mockRepos := database.NewMockRepoStore()
	mockRepos.GetByIDsFunc.SetDefaultReturn([]*types.Repo{{ID: 1, Name: "repo1"}}, nil)
	mockDB.ReposFunc.SetDefaultReturn(mockRepos)

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.ReadFileFunc.SetDefaultReturn([]byte("test\nfirst\nfour\nlines\nplus\nsome\nmore"), nil)

	mockEmbeddingsClient := embeddings.NewMockClient()
	mockEmbeddingsClient.SearchFunc.SetDefaultReturn(&embeddings.EmbeddingCombinedSearchResults{
		CodeResults: embeddings.EmbeddingSearchResults{{
			FileName:  "testfile",
			StartLine: 0,
			EndLine:   4,
		}},
	}, nil)

	repoEmbeddingJobsStore := repo.NewMockRepoEmbeddingJobsStore()
	contextDetectionJobsStore := contextdetection.NewMockContextDetectionEmbeddingJobsStore()

	resolver := NewResolver(
		mockDB,
		logger,
		mockGitserver,
		mockEmbeddingsClient,
		repoEmbeddingJobsStore,
		contextDetectionJobsStore,
	)

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			Embeddings:  &schema.Embeddings{Enabled: true},
			Completions: &schema.Completions{Enabled: true},
		},
	})

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
	ffs := featureflag.NewMemoryStore(map[string]bool{"cody-experimental": true}, nil, nil)
	ctx = featureflag.WithFlags(ctx, ffs)

	results, err := resolver.EmbeddingsSearch(ctx, graphqlbackend.EmbeddingsSearchInputArgs{
		Repos:            []graphql.ID{graphqlbackend.MarshalRepositoryID(3)},
		Query:            "test",
		CodeResultsCount: 1,
		TextResultsCount: 1,
	})
	require.NoError(t, err)

	codeResults, err := results.CodeResults(ctx)
	require.NoError(t, err)
	require.Equal(t, "test\nfirst\nfour\nlines", codeResults[0].Content(ctx))

}

func Test_extractLineRange(t *testing.T) {
	cases := []struct {
		input      []byte
		start, end int
		output     []byte
	}{{
		[]byte("zero\none\ntwo\nthree\n"),
		0, 2,
		[]byte("zero\none"),
	}, {
		[]byte("zero\none\ntwo\nthree\n"),
		1, 2,
		[]byte("one"),
	}, {
		[]byte("zero\none\ntwo\nthree\n"),
		1, 2,
		[]byte("one"),
	}, {
		[]byte(""),
		1, 2,
		[]byte(""),
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := extractLineRange(tc.input, tc.start, tc.end)
			require.Equal(t, tc.output, got)
		})
	}
}
