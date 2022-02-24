package client

import (
	"context"

	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/execute"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SearchClient interface {
	Plan(
		ctx context.Context,
		db database.DB,
		version string,
		patternType *string,
		searchQuery string,
		stream streaming.Sender,
		settings *schema.Settings,
	) (*run.SearchInputs, error)

	Execute(
		ctx context.Context,
		db database.DB,
		stream streaming.Sender,
		inputs *run.SearchInputs,
	) (_ *search.Alert, err error)
}

func NewSearchClient(zoektStreamer zoekt.Streamer, searcherURLs *endpoint.Map) *searchClient {
	return &searchClient{
		zoekt:        zoektStreamer,
		searcherURLs: searcherURLs,
	}
}

type searchClient struct {
	zoekt        zoekt.Streamer
	searcherURLs *endpoint.Map
}

func (s *searchClient) Plan(
	ctx context.Context,
	db database.DB,
	version string,
	patternType *string,
	searchQuery string,
	stream streaming.Sender,
	settings *schema.Settings,
) (*run.SearchInputs, error) {
	return run.NewSearchInputs(ctx, db, version, patternType, searchQuery, stream, settings)
}

func (s *searchClient) Execute(
	ctx context.Context,
	db database.DB,
	stream streaming.Sender,
	inputs *run.SearchInputs,
) (*search.Alert, error) {
	jobArgs := &job.Args{
		SearchInputs:        inputs,
		Zoekt:               s.zoekt,
		SearcherURLs:        s.searcherURLs,
		OnSourcegraphDotCom: envvar.SourcegraphDotComMode(),
	}
	return execute.Execute(ctx, db, stream, jobArgs)
}
