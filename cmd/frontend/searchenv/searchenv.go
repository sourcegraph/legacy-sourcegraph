package searchenv

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	searcherURL = env.Get("SEARCHER_URL", "k8s+http://searcher:3181", "searcher server URL")

	searcherURLsOnce sync.Once
	searcherURLs     *endpoint.Map

	indexedEndpointsOnce sync.Once
	indexedEndpoints     *endpoint.Map

	IndexedListTTL = func() time.Duration {
		ttl, _ := time.ParseDuration(env.Get("SRC_INDEXED_SEARCH_LIST_CACHE_TTL", "", "Indexed search list cache TTL"))
		if ttl == 0 {
			if envvar.SourcegraphDotComMode() {
				ttl = 30 * time.Second
			} else {
				ttl = 5 * time.Second
			}
		}
		return ttl
	}()
)

func SearcherURLs() *endpoint.Map {
	searcherURLsOnce.Do(func() {
		if len(strings.Fields(searcherURL)) == 0 {
			searcherURLs = endpoint.Empty(errors.New("a searcher service has not been configured"))
		} else {
			searcherURLs = endpoint.New(searcherURL)
		}
	})
	return searcherURLs
}

func IndexedEndpoints() *endpoint.Map {
	indexedEndpointsOnce.Do(func() {
		if addr := zoektAddr(os.Environ()); addr != "" {
			indexedEndpoints = endpoint.New(addr)
		}
	})
	return indexedEndpoints
}

func zoektAddr(environ []string) string {
	if addr, ok := getEnv(environ, "INDEXED_SEARCH_SERVERS"); ok {
		return addr
	}

	// Backwards compatibility: We used to call this variable ZOEKT_HOST
	if addr, ok := getEnv(environ, "ZOEKT_HOST"); ok {
		return addr
	}

	// Not set, use the default (service discovery on the indexed-search
	// statefulset)
	return "k8s+rpc://indexed-search:6070?kind=sts"
}

func getEnv(environ []string, key string) (string, bool) {
	key = key + "="
	for _, env := range environ {
		if strings.HasPrefix(env, key) {
			return env[len(key):], true
		}
	}
	return "", false
}
