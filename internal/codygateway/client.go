package codygateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type LimitStatus struct {
	// Feature is not part of the returned JSON.
	Feature Feature

	IntervalLimit int64      `json:"limit"`
	IntervalUsage int64      `json:"usage"`
	TimeInterval  string     `json:"interval"`
	Expiry        *time.Time `json:"expiry"`
}

func (rl LimitStatus) PercentUsed() int {
	if rl.IntervalLimit == 0 {
		return 100
	}
	return int(math.Ceil(float64(rl.IntervalUsage) / float64(rl.IntervalLimit) * 100))
}

type Attribution struct {
	Repositories []string
	LimitHit     bool
}

type Client interface {
	GetLimits(ctx context.Context) ([]LimitStatus, error)
	Attribution(ctx context.Context, snippet string, limit int) (Attribution, error)
}

func NewClientFromSiteConfig(cli httpcli.Doer) (_ Client, ok bool) {
	config := conf.Get().SiteConfig()
	cc := conf.GetCompletionsConfig(config)
	ec := conf.GetEmbeddingsConfig(config)

	// If neither completions nor embeddings are configured, return empty.
	if cc == nil && ec == nil {
		return nil, false
	}

	// TODO: What if customer is BYOK? How do we talk to gateway then?
	// If neither completions nor embeddings use Cody Gateway, return empty.
	ccUsingGateway := cc != nil && cc.Provider == conftypes.CompletionsProviderNameSourcegraph
	ecUsingGateway := ec != nil && ec.Provider == conftypes.EmbeddingsProviderNameSourcegraph
	if !ccUsingGateway && !ecUsingGateway {
		return nil, false
	}

	// It's possible the user is only using Cody Gateway for completions _or_ embeddings
	// make sure to get the url/token for the sourcegraph provider
	// start with the embeddings since there are fewer options
	endpoint := ec.Endpoint
	token := ec.AccessToken
	if ec.Provider != conftypes.EmbeddingsProviderNameSourcegraph {
		endpoint = cc.Endpoint
		token = cc.AccessToken
	}

	return NewClient(cli, endpoint, token), true
}

func NewClient(cli httpcli.Doer, endpoint string, accessToken string) Client {
	return &client{
		cli:         cli,
		endpoint:    endpoint,
		accessToken: accessToken,
	}
}

type client struct {
	cli         httpcli.Doer
	endpoint    string
	accessToken string
}

func (c *client) GetLimits(ctx context.Context) ([]LimitStatus, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, err
	}
	u.Path = "v1/limits"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("request failed with status: %d", errors.Safe(resp.StatusCode))
	}

	var featureLimits map[string]LimitStatus
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&featureLimits); err != nil {
		return nil, err
	}

	rateLimits := make([]LimitStatus, 0, len(featureLimits))
	for f, limit := range featureLimits {
		feat := Feature(f)
		// Check if this is a limit for a feature we know about.
		if feat.IsValid() {
			limit.Feature = feat
			rateLimits = append(rateLimits, limit)
		}
	}

	// Make sure the limits are always returned in the same order, since the map
	// above doesn't have deterministic ordering.
	sort.Slice(rateLimits, func(i, j int) bool {
		return rateLimits[i].Feature < rateLimits[j].Feature
	})

	return rateLimits, nil
}

func (c *client) Attribution(ctx context.Context, snippet string, limit int) (Attribution, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return Attribution{}, err
	}
	u.Path = "v1/attribution"
	body := new(bytes.Buffer)
	// TODO: Use the same struct as gateway
	if err := json.NewEncoder(body).Encode(struct {
		Snippet string `json:"snippet"`
		Limit   int    `json:"limit"`
	}{
		Snippet: snippet,
		Limit:   limit,
	}); err != nil {
		return Attribution{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body.Bytes()))
	if err != nil {
		return Attribution{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	resp, err := c.cli.Do(req)
	if err != nil {
		return Attribution{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Attribution{}, errors.Newf("request failed with status: %d", errors.Safe(resp.StatusCode))
	}
	// TODO: Use the same struct as gateway
	var gatewayResponse struct {
		Repositories []struct {
			Name string `json:"name"`
		} `json:"repositories"`
		TotalCount int  `json:"totalCount,omitempty"`
		LimitHit   bool `json:"limitHit,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gatewayResponse); err != nil {
		return Attribution{}, errors.Wrap(err, "cannot interpret gateway response")
	}
	a := Attribution{
		Repositories: make([]string, len(gatewayResponse.Repositories)),
		LimitHit:     gatewayResponse.LimitHit,
	}
	for i, r := range gatewayResponse.Repositories {
		a.Repositories[i] = r.Name
	}
	return a, nil
}
