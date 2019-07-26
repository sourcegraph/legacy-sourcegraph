package githuboauth

import (
	"fmt"
	"net/url"

	"github.com/dghubble/gologin/github"
	"golang.org/x/oauth2"
	"sourcegraph.com/cmd/frontend/envvar"
	"sourcegraph.com/enterprise/cmd/frontend/auth/oauth"
	"sourcegraph.com/pkg/extsvc"
	gh "sourcegraph.com/pkg/extsvc/github"
	"sourcegraph.com/schema"
)

const sessionKey = "githuboauth@0"

func parseProvider(p *schema.GitHubAuthProvider, sourceCfg schema.AuthProviders) (provider *oauth.Provider, problems []string) {
	rawURL := p.Url
	if rawURL == "" {
		rawURL = "https://github.com/"
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		problems = append(problems, fmt.Sprintf("Could not parse GitHub URL %q. You will not be able to login via this GitHub instance.", rawURL))
		return nil, problems
	}
	codeHost := extsvc.NewCodeHost(parsedURL, gh.ServiceType)
	oauth2Cfg := oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       requestedScopes(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  codeHost.BaseURL.ResolveReference(&url.URL{Path: "/login/oauth/authorize"}).String(),
			TokenURL: codeHost.BaseURL.ResolveReference(&url.URL{Path: "/login/oauth/access_token"}).String(),
		},
	}
	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix:   authPrefix,
		OAuth2Config: oauth2Cfg,
		SourceConfig: sourceCfg,
		StateConfig:  getStateConfig(),
		ServiceID:    codeHost.ServiceID,
		ServiceType:  codeHost.ServiceType,
		Login:        github.LoginHandler(&oauth2Cfg, nil),
		Callback: github.CallbackHandler(
			&oauth2Cfg,
			oauth.SessionIssuer(&sessionIssuerHelper{
				CodeHost:    codeHost,
				clientID:    p.ClientID,
				allowSignup: p.AllowSignup,
			}, sessionKey),
			nil,
		),
	}), nil
}

func requestedScopes() []string {
	if envvar.SourcegraphDotComMode() {
		return []string{"public_repo", "user:email"}
	}
	return []string{"repo", "user:email"}
}
