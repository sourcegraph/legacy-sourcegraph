package gitserver

import (
	"crypto/md5"
	"encoding/binary"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

var addrForRepoInvoked = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_invoked",
	Help: "Number of times gitserver.AddrForRepo was invoked",
}, []string{"user_agent"})

func NewGitserverAddressesFromConf() GitServerAddresses {
	cfg := conf.Get()
	addrs := GitServerAddresses{
		Addresses: cfg.ServiceConnectionConfig.GitServers,
	}
	if cfg.ExperimentalFeatures != nil {
		addrs.PinnedServers = cfg.ExperimentalFeatures.GitServerPinnedRepos
	}
	return addrs
}

func newTestGitserverAddresses(addrs []string) GitServerAddresses {
	return GitServerAddresses{
		Addresses: addrs,
	}
}

type GitServerAddresses struct {
	Addresses     []string
	PinnedServers map[string]string
}

func (g GitServerAddresses) AddrForRepo(userAgent string, repo api.RepoName) string {
	repo = protocol.NormalizeRepo(repo) // in case the caller didn't already normalize it
	rs := string(repo)

	if pinnedAddr, ok := g.PinnedServers[rs]; ok {
		return pinnedAddr
	}

	return addrForKey(rs, g.Addresses)
}

// addrForKey returns the gitserver address to use for the given string key,
// which is hashed for sharding purposes.
func addrForKey(key string, addrs []string) string {
	sum := md5.Sum([]byte(key))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(addrs))
	return addrs[serverIndex]
}
