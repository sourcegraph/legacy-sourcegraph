package conf

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/schema"
)

// AuthProviderType returns the type string for the auth provider.
func AuthProviderType(p schema.AuthProviders) string {
	switch {
	case p.Builtin != nil:
		return p.Builtin.Type
	case p.Openidconnect != nil:
		return p.Openidconnect.Type
	case p.Saml != nil:
		return p.Saml.Type
	case p.HttpHeader != nil:
		return p.HttpHeader.Type
	case p.Github != nil:
		return p.Github.Type
	case p.Gitlab != nil:
		return p.Gitlab.Type
	default:
		return ""
	}
}

// AuthPublic reports whether the site is public. This leads to a significantly degraded user
// experience, and as a result, is currently only supported on Sourcegraph.com.
func AuthPublic() bool { return authPublic(Get()) }
func authPublic(c *Unified) bool {
	return envvar.SourcegraphDotComMode()
}

// AuthAllowSignup reports whether the site allows signup. Currently only the builtin auth provider
// allows signup. AuthAllowSignup returns true if auth.providers' builtin provider has allowSignup
// true (in site config).
func AuthAllowSignup() bool { return authAllowSignup(Get()) }
func authAllowSignup(c *Unified) bool {
	for _, p := range c.Critical.AuthProviders {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return false
}
