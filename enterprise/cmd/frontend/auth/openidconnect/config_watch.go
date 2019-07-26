package openidconnect

import (
	"context"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/cmd/frontend/auth/providers"
	"sourcegraph.com/pkg/conf"
	"sourcegraph.com/schema"
)

func getProviders() []providers.Provider {
	var cfgs []*schema.OpenIDConnectAuthProvider
	for _, p := range conf.Get().Critical.AuthProviders {
		if p.Openidconnect == nil {
			continue
		}
		cfgs = append(cfgs, p.Openidconnect)
	}
	ps := make([]providers.Provider, 0, len(cfgs))
	for _, cfg := range cfgs {
		p := &provider{config: *cfg}
		ps = append(ps, p)
	}
	return ps
}

func init() {
	go func() {
		conf.Watch(func() {
			ps := getProviders()
			for _, p := range ps {
				go func(p providers.Provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching OpenID Connect service provider metadata.", "error", err)
					}
				}(p)
			}
			providers.Update("openidconnect", ps)
		})
	}()
}
