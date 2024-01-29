package guardrails

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/attribution"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	resolver := &resolvers.GuardrailsResolver{}
	if envvar.SourcegraphDotComMode() {
		// On DotCom guardrails endpoint runs search, and is initialized at startup.
		searchClient := client.New(observationCtx.Logger, db, gitserver.NewClient("http.guardrails.search"))
		service := attribution.NewLocalSearch(observationCtx, searchClient)
		resolver.AttributionService = func() attribution.Service { return service }
	} else {
		// On an Enterprise instance endpoint proxies to gateway, and is re-initialized
		// in case site-config changes.
		initLogic := &enterpriseInitialization{observationCtx: observationCtx}
		resolver.AttributionService = func() attribution.Service { return initLogic.Service() }
	}
	enterpriseServices.GuardrailsResolver = resolver
	return nil
}

type enterpriseInitialization struct {
	observationCtx *observation.Context
	mu             sync.Mutex
	client         codygateway.Client
	endpoint       string
	token          string
}

func (e *enterpriseInitialization) Service() attribution.Service {
	e.mu.Lock()
	defer e.mu.Unlock()
	endpoint, token := e.newConfig()
	if e.endpoint != endpoint || e.token != token {
		e.endpoint = endpoint
		e.token = token
		e.client = codygateway.NewClient(httpcli.ExternalDoer, endpoint, token)
	}
	return attribution.NewGatewayProxy(e.observationCtx, e.client)
}

func (e *enterpriseInitialization) newConfig() (string, string) {
	config := conf.Get().SiteConfig()
	// Explicit attribution gateway config overrides autocomplete config (if used).
	if gateway := conf.GetAttributionGateway(config); gateway != nil {
		return gateway.Endpoint, gateway.AccessToken
	}
	// Fall back to autocomplete config if no explicit gateway config.
	cc := conf.GetCompletionsConfig(config)
	ccUsingGateway := cc != nil && cc.Provider == conftypes.CompletionsProviderNameSourcegraph
	if ccUsingGateway {
		return cc.Endpoint, cc.AccessToken
	}
	return "", ""
}

func alwaysAllowed(context.Context, string) (bool, error) {
	return true, nil
}

func NewAttributionTest(observationCtx *observation.Context) func(context.Context, string) (bool, error) {
	// Attribution is only-enterprise, dotcom lets everything through.
	if envvar.SourcegraphDotComMode() {
		return alwaysAllowed
	}
	initLogic := &enterpriseInitialization{observationCtx: observationCtx}
	return func(ctx context.Context, snippet string) (bool, error) {
		// Check if attribution is on, permit everything if it's off.
		c := conf.GetConfigFeatures(conf.Get().SiteConfig())
		if !c.Attribution {
			return true, nil
		}
		// Attribution not available. Mode is permissive.
		attribution, err := initLogic.Service().SnippetAttribution(ctx, snippet, 1)
		// Attribution not available. Mode is permissive.
		if err != nil {
			return true, err
		}
		// Permit completion if no attribution found.
		return len(attribution.RepositoryNames) == 0, nil
	}
}
