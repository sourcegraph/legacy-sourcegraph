package gerrit

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init() {
	const pkgName = "gerrit"
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := parseConfig(cfg)
		return problems
	})

	go conf.Watch(func() {
		newProviders, _ := parseConfig(conf.Get())
		fmt.Println(newProviders)
		newProviderList := make([]providers.Provider, len(newProviders))
		for i, p := range newProviders {
			p := p // capture loop variable
			newProviderList[i] = &p
		}
		providers.Update(pkgName, newProviderList)
	})
}

type Provider struct {
	ServiceID   string
	ServiceType string
}

func parseConfig(cfg conftypes.SiteConfigQuerier) (ps []Provider, problems conf.Problems) {
	for _, pr := range cfg.SiteConfig().AuthProviders {
		if pr.Gerrit == nil {
			continue
		}

		provider, providerProblems := parseProvider(pr.Gerrit)
		problems = append(problems, conf.NewSiteProblems(providerProblems...)...)
		ps = append(ps, provider)
	}

	return ps, problems
}

func parseProvider(p *schema.GerritAuthProvider) (Provider, []string) {
	return Provider{
		ServiceID:   p.Url,
		ServiceType: p.Type,
	}, []string{}
}

func (p *Provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		Type: "gerrit",
		ID:   "http://localhost:8080/",
	}
}

func (p *Provider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		Gerrit: &schema.GerritAuthProvider{
			Type: p.ServiceType,
			Url:  p.ServiceID,
		},
	}
}

func (p *Provider) CachedInfo() *providers.Info {
	return &providers.Info{
		ServiceID:         p.ServiceID,
		ClientID:          "",
		DisplayName:       "Gerrit",
		AuthenticationURL: p.ServiceID,
	}
}

func (p *Provider) Refresh(ctx context.Context) error {
	return nil
}

func (p *Provider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	return gerrit.GetPublicExternalAccountData(ctx, &account.AccountData)
}
