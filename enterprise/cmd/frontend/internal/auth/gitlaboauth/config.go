package gitlaboauth

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/schema"
)

const PkgName = "gitlaboauth"

func Init(db dbutil.DB) {
	conf.ContributeValidator(func(cfg conf.Unified) conf.Problems {
		_, problems := parseConfig(db, &cfg)
		return problems
	})
	goroutine.Go(func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(db, conf.Get())
			if len(newProviders) == 0 {
				providers.Update(PkgName, nil)
			} else {
				newProvidersList := make([]providers.Provider, 0, len(newProviders))
				for _, p := range newProviders {
					newProvidersList = append(newProvidersList, p)
				}
				providers.Update(PkgName, newProvidersList)
			}
		})
	})
}

func parseConfig(db dbutil.DB, cfg *conf.Unified) (ps map[schema.GitLabAuthProvider]providers.Provider, problems conf.Problems) {
	ps = make(map[schema.GitLabAuthProvider]providers.Provider)
	for _, pr := range cfg.AuthProviders {
		if pr.Gitlab == nil {
			continue
		}

		if cfg.ExternalURL == "" {
			problems = append(problems, conf.NewSiteProblem("`externalURL` was empty and it is needed to determine the OAuth callback URL."))
			continue
		}
		externalURL, err := url.Parse(cfg.ExternalURL)
		if err != nil {
			problems = append(problems, conf.NewSiteProblem("Could not parse `externalURL`, which is needed to determine the OAuth callback URL."))
			continue
		}
		callbackURL := *externalURL
		callbackURL.Path = "/.auth/gitlab/callback"

		provider, providerMessages := parseProvider(db, callbackURL.String(), pr.Gitlab, pr)
		problems = append(problems, conf.NewSiteProblems(providerMessages...)...)
		if provider != nil {
			ps[*pr.Gitlab] = provider
		}
	}
	return ps, problems
}
