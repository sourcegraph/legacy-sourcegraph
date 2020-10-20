package campaigns

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// SupportedExternalServices are the external service types currently supported
// by the campaigns feature. Repos that are associated with external services
// whose type is not in this list will simply be filtered out from the search
// results.
var SupportedExternalServices = map[string]struct{}{
	extsvc.TypeGitHub:          {},
	extsvc.TypeBitbucketServer: {},
	extsvc.TypeGitLab:          {},
}

// IsRepoSupported returns whether the given ExternalRepoSpec is supported by
// the campaigns feature, based on the external service type.
func IsRepoSupported(spec *api.ExternalRepoSpec) bool {
	_, ok := SupportedExternalServices[spec.ServiceType]
	return ok
}

// IsKindSupported returns whether the given extsvc Kind is supported by
// campaigns.
func IsKindSupported(extSvcKind string) bool {
	_, ok := SupportedExternalServices[extsvc.KindToType(extSvcKind)]
	return ok
}

// Keyer represents items that return a unique key
type Keyer interface {
	Key() string
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
