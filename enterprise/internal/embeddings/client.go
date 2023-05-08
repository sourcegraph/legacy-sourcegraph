package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func defaultEndpoints() *endpoint.Map {
	return endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
		return conns.Embeddings
	})
}

var defaultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternalClientFactory("embeddings").Doer()
	if err != nil {
		panic(err)
	}
	return d
}()

func NewClient() *Client {
	return &Client{
		Endpoints:  defaultEndpoints(),
		HTTPClient: defaultDoer,
	}
}

type Client struct {
	// Endpoints to embeddings service.
	Endpoints *endpoint.Map

	// HTTP client to use
	HTTPClient httpcli.Doer
}

type EmbeddingsSearchParameters struct {
	RepoName         api.RepoName `json:"repoName"`
	RepoID           api.RepoID   `json:"repoID"`
	Query            string       `json:"query"`
	CodeResultsCount int          `json:"codeResultsCount"`
	TextResultsCount int          `json:"textResultsCount"`

	UseDocumentRanks bool `json:"useDocumentRanks"`
}

type EmbeddingsMultiSearchParameters struct {
	RepoNames        []api.RepoName `json:"repoNames"`
	RepoIDs          []api.RepoID   `json:"repoIDs"`
	Query            string         `json:"query"`
	CodeResultsCount int            `json:"codeResultsCount"`
	TextResultsCount int            `json:"textResultsCount"`

	UseDocumentRanks bool `json:"useDocumentRanks"`
	// If set to "True", EmbeddingSearchResult.Debug will contain useful information about scoring.
	Debug bool `json:"debug"`
}

type IsContextRequiredForChatQueryParameters struct {
	Query string `json:"query"`
}

type IsContextRequiredForChatQueryResult struct {
	IsRequired bool `json:"isRequired"`
}

func (c *Client) Search(ctx context.Context, args EmbeddingsSearchParameters) (*EmbeddingSearchResults, error) {
	url, err := c.url(args.RepoName)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpPost(ctx, "search", url, args)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, errors.Errorf(
			"Embeddings.Search http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var response EmbeddingSearchResults
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) MultiSearch(ctx context.Context, args EmbeddingsMultiSearchParameters) (*EmbeddingSearchResults, error) {
	partitions, err := c.urls(args.RepoNames, args.RepoIDs)
	if err != nil {
		return nil, err
	}

	p := pool.NewWithResults[*EmbeddingSearchResults]().WithContext(ctx)
	for endpoint, partition := range partitions {
		endpoint := endpoint

		// make a copy for this request
		args := args
		args.RepoNames = partition.repoNames
		args.RepoIDs = partition.repoIDs

		p.Go(func(ctx context.Context) (*EmbeddingSearchResults, error) {
			resp, err := c.httpPost(ctx, "multiSearch", endpoint, args)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				// best-effort inclusion of body in error message
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
				return nil, errors.Errorf(
					"Embeddings.Search http status %d: %s",
					resp.StatusCode,
					string(body),
				)
			}

			var response EmbeddingSearchResults
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return nil, err
			}
			return &response, nil
		})
	}

	allResults, err := p.Wait()
	if err != nil {
		return nil, err
	}

	var combinedResult EmbeddingSearchResults
	for _, result := range allResults {
		combinedResult.CodeResults = mergeSearchResults(combinedResult.CodeResults, result.CodeResults, args.CodeResultsCount)
		combinedResult.TextResults = mergeSearchResults(combinedResult.TextResults, result.TextResults, args.TextResultsCount)
	}

	return &combinedResult, nil
}

func mergeSearchResults(a, b []EmbeddingSearchResult, max int) []EmbeddingSearchResult {
	merged := append(a, b...)
	sort.Slice(merged, func(i, j int) bool { return merged[i].Score() > merged[j].Score() })
	return merged[:max]
}

func (c *Client) IsContextRequiredForChatQuery(ctx context.Context, args IsContextRequiredForChatQueryParameters) (bool, error) {
	resp, err := c.httpPost(ctx, "isContextRequiredForChatQuery", "", args)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return false, errors.Errorf(
			"Embeddings.IsContextRequiredForChatQuery http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var response IsContextRequiredForChatQueryResult
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, err
	}
	return response.IsRequired, nil
}

func (c *Client) url(repo api.RepoName) (string, error) {
	if c.Endpoints == nil {
		return "", errors.New("an embeddings service has not been configured")
	}
	return c.Endpoints.Get(string(repo))
}

type repoPartition struct {
	repoNames []api.RepoName
	repoIDs   []api.RepoID
}

// returns a map from URL to a list of indexes into the input list of repos
func (c *Client) urls(repos []api.RepoName, repoIDs []api.RepoID) (map[string]repoPartition, error) {
	if c.Endpoints == nil {
		return nil, errors.New("an embeddings service has not been configured")
	}

	repoStrings := make([]string, len(repos))
	for i, repo := range repos {
		repoStrings[i] = string(repo)
	}

	endpoints, err := c.Endpoints.GetMany(repoStrings...)
	if err != nil {
		return nil, err
	}

	res := make(map[string]repoPartition)
	for i, endpoint := range endpoints {
		res[endpoint] = repoPartition{
			repoNames: append(res[endpoint].repoNames, repos[i]),
			repoIDs:   append(res[endpoint].repoIDs, repoIDs[i]),
		}
	}
	return res, nil
}

func (c *Client) httpPost(
	ctx context.Context,
	method string,
	url string,
	payload any,
) (resp *http.Response, err error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	req, err := http.NewRequest("POST", url+method, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	return c.HTTPClient.Do(req)
}
