// Package hooks allow hooking into the frontend.
package hooks

import (
	"net/http"
)

// PostAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is authenticated when PostAuthMiddleware is called.
var PostAuthMiddleware func(http.Handler) http.Handler

// LicenseInfo contains information about the legitimate usage of the current
// license on the instance.
type LicenseInfo struct {
	CodeScaleLimit         string
	CodeScaleCloseToLimit  bool
	CodeScaleExceededLimit bool
}

var GetLicenseInfo = func(isSiteAdmin bool) *LicenseInfo { return nil }
