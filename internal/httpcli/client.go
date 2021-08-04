package httpcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/cockroachdb/errors"
	"github.com/gregjones/httpcache"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// A Doer captures the Do method of an http.Client. It facilitates decorating
// an http.Client with orthogonal concerns such as logging, metrics, retries,
// etc.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// DoerFunc is function adapter that implements the http.RoundTripper
// interface by calling itself.
type DoerFunc func(*http.Request) (*http.Response, error)

// Do implements the Doer interface.
func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// A Middleware function wraps a Doer with a layer of behaviour. It's used
// to decorate an http.Client with orthogonal layers of behaviour such as
// logging, instrumentation, retries, etc.
type Middleware func(Doer) Doer

// NewMiddleware returns a Middleware stack composed of the given Middlewares.
func NewMiddleware(mws ...Middleware) Middleware {
	return func(bottom Doer) (stacked Doer) {
		stacked = bottom
		for _, mw := range mws {
			stacked = mw(stacked)
		}
		return stacked
	}
}

// A Opt configures an aspect of a given *http.Client,
// returning an error in case of failure.
type Opt func(*http.Client) error

// A Factory constructs an http.Client with the given functional
// options applied, returning an aggregate error of the errors returned by
// all those options.
type Factory struct {
	stack  Middleware
	common []Opt
}

// redisCache is a HTTP cache backed by Redis. The TTL of a week is a balance
// between caching values for a useful amount of time versus growing the cache
// too large.
var redisCache = rcache.NewWithTTL("http", 604800)

// ExternalClientFactory is a httpcli.Factory with common options
// and middleware pre-set for communicating with external services.
var ExternalClientFactory = NewExternalClientFactory()

// NewExternalClientFactory returns a httpcli.Factory with common options
// and middleware pre-set for communicating with external services.
func NewExternalClientFactory() *Factory {
	return NewFactory(
		NewMiddleware(
			ContextErrorMiddleware,
		),
		NewTimeoutOpt(5*time.Minute),
		// ExternalTransportOpt needs to be before TracedTransportOpt and
		// NewCachedTransportOpt since it wants to extract a http.Transport,
		// not a generic http.RoundTripper.
		ExternalTransportOpt,
		NewErrorResilientTransportOpt(
			NewRetryPolicy(MaxRetries()),
			rehttp.ExpJitterDelay(200*time.Millisecond, 10*time.Second),
		),
		TracedTransportOpt,
		NewCachedTransportOpt(redisCache, true),
	)
}

// ExternalDoer is a shared client for external communication. This is a
// convenience for existing uses of http.DefaultClient.
var ExternalDoer, _ = ExternalClientFactory.Doer()

// ExternalClient returns a shared client for external communication. This is
// a convenience for existing uses of http.DefaultClient.
var ExternalClient, _ = ExternalClientFactory.Client()

// InternalClientFactory is a httpcli.Factory with common options
// and middleware pre-set for communicating with internal services.
var InternalClientFactory = NewInternalClientFactory("internal")

// NewInternalClientFactory returns a httpcli.Factory with common options
// and middleware pre-set for communicating with internal services.
func NewInternalClientFactory(subsystem string) *Factory {
	return NewFactory(
		NewMiddleware(
			ContextErrorMiddleware,
		),
		NewTimeoutOpt(time.Minute), // Needs to be high for retries.
		NewMaxIdleConnsPerHostOpt(500),
		NewErrorResilientTransportOpt(
			NewRetryPolicy(MaxRetries()),
			rehttp.ExpJitterDelay(50*time.Millisecond, 5*time.Second),
		),
		MeteredTransportOpt(subsystem),
		TracedTransportOpt,
	)
}

// InternalDoer is a shared client for external communication. This is a
// convenience for existing uses of http.DefaultClient.
var InternalDoer, _ = InternalClientFactory.Doer()

// InternalClient returns a shared client for external communication. This is
// a convenience for existing uses of http.DefaultClient.
var InternalClient, _ = InternalClientFactory.Client()

// Doer returns a new Doer wrapped with the middleware stack
// provided in the Factory constructor and with the given common
// and base opts applied to it.
func (f Factory) Doer(base ...Opt) (Doer, error) {
	cli, err := f.Client(base...)
	if err != nil {
		return nil, err
	}

	if f.stack != nil {
		return f.stack(cli), nil
	}

	return cli, nil
}

// Client returns a new http.Client configured with the
// given common and base opts, but not wrapped with any
// middleware.
func (f Factory) Client(base ...Opt) (*http.Client, error) {
	opts := make([]Opt, 0, len(f.common)+len(base))
	opts = append(opts, base...)
	opts = append(opts, f.common...)

	var cli http.Client
	var err *multierror.Error

	for _, opt := range opts {
		err = multierror.Append(err, opt(&cli))
	}

	return &cli, err.ErrorOrNil()
}

// NewFactory returns a Factory that applies the given common
// Opts after the ones provided on each invocation of Client or Doer.
//
// If the given Middleware stack is not nil, the final configured client
// will be wrapped by it before being returned from a call to Doer, but not Client.
func NewFactory(stack Middleware, common ...Opt) *Factory {
	return &Factory{stack: stack, common: common}
}

//
// Common Middleware
//

// HeadersMiddleware returns a middleware that wraps a Doer
// and sets the given headers.
func HeadersMiddleware(headers ...string) Middleware {
	if len(headers)%2 != 0 {
		panic("missing header values")
	}
	return func(cli Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			for i := 0; i < len(headers); i += 2 {
				req.Header.Add(headers[i], headers[i+1])
			}
			return cli.Do(req)
		})
	}
}

// ContextErrorMiddleware wraps a Doer with context.Context error
// handling.  It checks if the request context is done, and if so,
// returns its error. Otherwise it returns the error from the inner
// Doer call.
func ContextErrorMiddleware(cli Doer) Doer {
	return DoerFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := cli.Do(req)
		if err != nil {
			// If we got an error, and the context has been canceled,
			// the context's error is probably more useful.
			if e := req.Context().Err(); e != nil {
				err = e
			}
		}
		return resp, err
	})
}

// GitHubProxyRedirectMiddleware rewrites requests to the "github-proxy" host
// to "https://api.github.com".
func GitHubProxyRedirectMiddleware(cli Doer) Doer {
	return DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() == "github-proxy" {
			req.URL.Host = "api.github.com"
			req.URL.Scheme = "https"
		}
		return cli.Do(req)
	})
}

//
// Common Opts
//

// ExternalTransportOpt returns an Opt that ensures the http.Client.Transport
// can contact non-Sourcegraph services. For example Admins can configure
// TLS/SSL settings.
func ExternalTransportOpt(cli *http.Client) error {
	tr, err := getTransportForMutation(cli)
	if err != nil {
		// TODO(keegancsmith) for now we don't support unwrappable
		// transports. https://github.com/sourcegraph/sourcegraph/pull/7741
		// https://github.com/sourcegraph/sourcegraph/pull/71
		if isUnwrappableTransport(cli) {
			return nil
		}
		return errors.Wrap(err, "httpcli.ExternalTransportOpt")
	}

	cli.Transport = &externalTransport{base: tr}
	return nil
}

func isUnwrappableTransport(cli *http.Client) bool {
	if cli.Transport == nil {
		return false
	}
	_, ok := cli.Transport.(interface{ UnwrappableTransport() })
	return ok
}

// NewCertPoolOpt returns a Opt that sets the RootCAs pool of an http.Client's
// transport.
func NewCertPoolOpt(certs ...string) Opt {
	return func(cli *http.Client) error {
		if len(certs) == 0 {
			return nil
		}

		tr, err := getTransportForMutation(cli)
		if err != nil {
			return errors.Wrap(err, "httpcli.NewCertPoolOpt")
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = new(tls.Config)
		}

		pool := x509.NewCertPool()
		tr.TLSClientConfig.RootCAs = pool

		for _, cert := range certs {
			if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
				return errors.New("httpcli.NewCertPoolOpt: invalid certificate")
			}
		}

		return nil
	}
}

// NewCachedTransportOpt returns an Opt that wraps the existing http.Transport
// of an http.Client with caching using the given Cache.
func NewCachedTransportOpt(c httpcache.Cache, markCachedResponses bool) Opt {
	return func(cli *http.Client) error {
		if cli.Transport == nil {
			cli.Transport = http.DefaultTransport
		}

		cli.Transport = &httpcache.Transport{
			Transport:           cli.Transport,
			Cache:               c,
			MarkCachedResponses: markCachedResponses,
		}

		return nil
	}
}

// TracedTransportOpt wraps an existing http.Transport of an http.Client with
// tracing functionality.
func TracedTransportOpt(cli *http.Client) error {
	if cli.Transport == nil {
		cli.Transport = http.DefaultTransport
	}

	cli.Transport = &ot.Transport{RoundTripper: cli.Transport}
	return nil
}

// MeteredTransportOpt returns an opt that wraps an existing http.Transport of a http.Client with
// metrics collection.
func MeteredTransportOpt(subsystem string) Opt {
	meter := metrics.NewRequestMeter(
		subsystem,
		"Total number of requests sent to "+subsystem,
	)

	return func(cli *http.Client) error {
		if cli.Transport == nil {
			cli.Transport = http.DefaultTransport
		}

		cli.Transport = meter.Transport(cli.Transport, func(u *url.URL) string {
			return u.Path
		})

		return nil
	}
}

// A regular expression to match the error returned by net/http when the
// configured number of redirects is exhausted. This error isn't typed
// specifically so we resort to matching on the error string.
var redirectsErrorRe = lazyregexp.New(`stopped after \d+ redirects\z`)

// A regular expression to match the error returned by net/http when the
// scheme specified in the URL is invalid. This error isn't typed
// specifically so we resort to matching on the error string.
var schemeErrorRe = lazyregexp.New(`unsupported protocol scheme`)

// A regular expression to match a DNSError no such host.
var noSuchHostErrorRe = lazyregexp.New(`no such host`)

// MaxRetries returns the max retries to be attempted, which should be passed
// to NewRetryPolicy. If we're in tests, it returns 1, otherwise it tries to
// parse SRC_HTTP_CLI_MAX_RETRIES and return that. If it can't, it defaults to 20.
func MaxRetries() int {
	if strings.HasSuffix(os.Args[0], ".test") {
		return 1
	}

	max, _ := strconv.Atoi(os.Getenv("SRC_HTTP_CLI_MAX_RETRIES"))
	if max == 0 {
		return 20
	}

	return max
}

// NewRetryPolicy returns a retry policy used in any Doer or Client returned
// by NewExternalClientFactory.
func NewRetryPolicy(max int) rehttp.RetryFn {
	return func(a rehttp.Attempt) (retry bool) {
		status := 0

		defer func() {
			if !retry {
				return
			}

			log15.Error(
				"retrying HTTP request",
				"attempt", a.Index,
				"method", a.Request.Method,
				"url", a.Request.URL,
				"status", status,
				"err", a.Error,
			)
		}()

		if a.Response != nil {
			status = a.Response.StatusCode
		}

		if a.Index > max { // Max retries
			return false
		}

		switch a.Error {
		case nil:
		case context.DeadlineExceeded, context.Canceled:
			return false
		default:
			if v, ok := a.Error.(*url.Error); ok {
				e := v.Error()
				// Don't retry if the error was due to too many redirects.
				if redirectsErrorRe.MatchString(e) {
					return false
				}

				// Don't retry if the error was due to an invalid protocol scheme.
				if schemeErrorRe.MatchString(e) {
					return false
				}

				// Don't retry more than 3 times for no such host errors.
				// This affords some resilience to dns unreliability while
				// preventing 20 attempts with a non existing name.
				if noSuchHostErrorRe.MatchString(e) && a.Index >= 3 {
					return false
				}

				// Don't retry if the error was due to TLS cert verification failure.
				if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
					return false
				}

			}
			// The error is likely recoverable so retry.
			return true
		}

		if status == 0 || status == 429 || (status >= 500 && status != 501) {
			return true
		}

		return false
	}
}

// NewErrorResilientTransportOpt returns an Opt that wraps an existing
// http.Transport of an http.Client with automatic retries.
func NewErrorResilientTransportOpt(retry rehttp.RetryFn, delay rehttp.DelayFn) Opt {
	return func(cli *http.Client) error {
		if cli.Transport == nil {
			cli.Transport = http.DefaultTransport
		}

		cli.Transport = rehttp.NewTransport(cli.Transport, retry, delay)
		return nil
	}
}

// NewIdleConnTimeoutOpt returns a Opt that sets the IdleConnTimeout of an
// http.Client's transport.
func NewIdleConnTimeoutOpt(timeout time.Duration) Opt {
	return func(cli *http.Client) error {
		tr, err := getTransportForMutation(cli)
		if err != nil {
			return errors.Wrap(err, "httpcli.NewIdleConnTimeoutOpt")
		}

		tr.IdleConnTimeout = timeout

		return nil
	}
}

// NewMaxIdleConnsPerHostOpt returns a Opt that sets the MaxIdleConnsPerHost field of an
// http.Client's transport.
func NewMaxIdleConnsPerHostOpt(max int) Opt {
	return func(cli *http.Client) error {
		tr, err := getTransportForMutation(cli)
		if err != nil {
			return errors.Wrap(err, "httpcli.NewMaxIdleConnsOpt")
		}

		tr.MaxIdleConnsPerHost = max

		return nil
	}
}

// NewTimeoutOpt returns a Opt that sets the Timeout field of an http.Client.
func NewTimeoutOpt(timeout time.Duration) Opt {
	return func(cli *http.Client) error {
		cli.Timeout = timeout
		return nil
	}
}

// getTransport returns the http.Transport for cli. If Transport is nil, it is
// set to a copy of the DefaultTransport. If it is the DefaultTransport, it is
// updated to a copy of the DefaultTransport.
//
// Use this function when you intend on mutating the transport.
func getTransportForMutation(cli *http.Client) (*http.Transport, error) {
	if cli.Transport == nil {
		cli.Transport = http.DefaultTransport
	}

	tr, ok := cli.Transport.(*http.Transport)
	if !ok {
		return nil, errors.Errorf("http.Client.Transport is not an *http.Transport: %T", cli.Transport)
	}

	tr = tr.Clone()
	cli.Transport = tr

	return tr, nil
}
