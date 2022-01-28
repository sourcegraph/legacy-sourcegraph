package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

var MockCountGoImporters func(ctx context.Context, repo api.RepoName) (int, error)

var (
	countGoImportersHTTPClient = httpcli.ExternalDoer
	goImportersCountCache      = rcache.NewWithTTL("go-importers-count", 14400) // 4 hours
)

// CountGoImporters returns the number of Go importers for the repository's Go subpackages. This is
// a special case used only on Sourcegraph.com for repository badges.
func CountGoImporters(ctx context.Context, cli httpcli.Doer, repo api.RepoName) (count int, err error) {
	if MockCountGoImporters != nil {
		return MockCountGoImporters(ctx, repo)
	}

	if !envvar.SourcegraphDotComMode() {
		// Avoid confusing users by exposing this on self-hosted instances, because it relies on the
		// public godoc.org API.
		return 0, errors.New("counting Go importers is not supported on self-hosted instances")
	}

	cacheKey := string(repo)
	b, ok := goImportersCountCache.Get(cacheKey)
	if ok {
		count, err = strconv.Atoi(string(b))
		if err == nil {
			return count, nil // cache hit
		}
		goImportersCountCache.Delete(cacheKey) // remove unexpectedly invalid cache value
	}

	defer func() {
		if err == nil {
			// Store in cache.
			goImportersCountCache.Set(cacheKey, []byte(strconv.Itoa(count)))
		}
	}()

	var q struct {
		Query     string
		Variables map[string]interface{}
	}

	q.Query = countGoImportersGraphQLQuery
	q.Variables = map[string]interface{}{
		"query": countGoImportersSearchQuery(repo),
	}

	body, err := json.Marshal(q)
	if err != nil {
		return 0, err
	}

	rawurl, err := gqlURL("CountGoImporters")
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", rawurl, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "ReadBody")
	}

	var v struct {
		Data struct {
			Search struct{ Results struct{ MatchCount int } }
		}
		Errors []interface{}
	}

	if err := json.Unmarshal(respBody, &v); err != nil {
		return 0, errors.Wrap(err, "Decode")
	}

	if len(v.Errors) > 0 {
		return 0, errors.Errorf("graphql: errors: %v", v.Errors)
	}

	return v.Data.Search.Results.MatchCount, nil
}

// gqlURL returns the frontend's internal GraphQL API URL, with the given ?queryName parameter
// which is used to keep track of the source and type of GraphQL queries.
func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}

func countGoImportersSearchQuery(repo api.RepoName) string {
	return strings.Join([]string{
		`type:file`,
		`f:go\.mod`,
		`patterntype:regexp`,
		`content:^\s+` + regexp.QuoteMeta(string(repo)) + `\S*\s+v\S`,
		`count:all`,
		`visibility:public`,
		`timeout:20s`,
	}, " ")
}

const countGoImportersGraphQLQuery = `
query CountGoImporters($query: String!) {
  search(query: $query) { results { matchCount } }
}`
