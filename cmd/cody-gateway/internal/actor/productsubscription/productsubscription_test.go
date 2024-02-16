package productsubscription

import (
	"testing"
	"time"

	"github.com/sourcegraph/log"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

func TestNewActor(t *testing.T) {
	concurrencyConfig := codygateway.ActorConcurrencyLimitConfig{
		Percentage: 50,
		Interval:   24 * time.Hour,
	}
	type args struct {
		s               dotcom.ProductSubscriptionState
		devLicensesOnly bool
	}
	tests := []struct {
		name        string
		args        args
		wantEnabled bool
	}{
		{
			name: "not dev only",
			args: args{
				dotcom.ProductSubscriptionState{
					CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: true,
							ChatCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsChatCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
							CodeCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsCodeCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
						},
					},
				},
				false,
			},
			wantEnabled: true,
		},
		{
			name: "dev only, not a dev license",
			args: args{
				dotcom.ProductSubscriptionState{
					CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: true,
							ChatCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsChatCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
							CodeCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsCodeCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
						},
					},
				},
				true,
			},
			wantEnabled: false,
		},
		{
			name: "dev only, is a dev license",
			args: args{
				dotcom.ProductSubscriptionState{
					CodyGatewayAccess: dotcom.ProductSubscriptionStateCodyGatewayAccess{
						CodyGatewayAccessFields: dotcom.CodyGatewayAccessFields{
							Enabled: true,
							ChatCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsChatCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
							CodeCompletionsRateLimit: &dotcom.CodyGatewayAccessFieldsCodeCompletionsRateLimitCodyGatewayRateLimit{
								RateLimitFields: dotcom.RateLimitFields{
									Limit:           10,
									IntervalSeconds: 10,
								},
							},
						},
					},
					ActiveLicense: &dotcom.ProductSubscriptionStateActiveLicenseProductLicense{
						Info: &dotcom.ProductSubscriptionStateActiveLicenseProductLicenseInfo{
							Tags: []string{"dev"},
						},
					},
				},
				true,
			},
			wantEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := newActor(nil, "", tt.args.s, tt.args.devLicensesOnly, concurrencyConfig)
			assert.Equal(t, act.AccessEnabled, tt.wantEnabled)
		})
	}
}

func TestGetSubscriptionAccountName(t *testing.T) {
	tests := []struct {
		name         string
		mockUsername string
		mockTags     []string
		wantName     string
	}{
		{
			name:         "has special license tag",
			mockUsername: "alice",
			mockTags:     []string{"trial", "customer:acme"},
			wantName:     "acme",
		},
		{
			name:         "use account username",
			mockUsername: "alice",
			mockTags:     []string{"plan:enterprise-1"},
			wantName:     "alice",
		},
		{
			name:     "no account name",
			wantName: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getSubscriptionAccountName(dotcom.ProductSubscriptionState{
				Account: &dotcom.ProductSubscriptionStateAccountUser{
					Username: test.mockUsername,
				},
				ActiveLicense: &dotcom.ProductSubscriptionStateActiveLicenseProductLicense{
					Info: &dotcom.ProductSubscriptionStateActiveLicenseProductLicenseInfo{
						Tags: test.mockTags,
					},
				},
			})
			assert.Equal(t, test.wantName, got)
		})
	}
}
func TestRemoveUnseenTokens(t *testing.T) {
	t.Run("removes unseen tokens", func(t *testing.T) {
		seen := map[string]struct{}{
			"slk_token1": {},
		}

		cache := &fakeListingCache{
			state: map[string][]byte{"v2:product-subscription:v1:slk_token1": nil, "v2:product-subscription:v2:slk_token3": nil},
		}

		removeUnseenTokens(seen, cache, log.Scoped("test"))
		assert.Equal(t, cache.calls, []struct{ call, key string }{{"ListAllKeys", ""}, {"Delete", "slk_token3"}})
	})

	t.Run("ignores malformed keys ", func(t *testing.T) {
		seen := map[string]struct{}{}

		cache := &fakeListingCache{
			state: map[string][]byte{"v2:product-subscription:": nil},
		}

		removeUnseenTokens(seen, cache, log.Scoped("test"))
		assert.Equal(t, cache.calls, []struct{ call, key string }{{"ListAllKeys", ""}})
	})
	t.Run("ignores malformed keys ", func(t *testing.T) {
		seen := map[string]struct{}{}

		cache := &fakeListingCache{
			state: map[string][]byte{"v2:product-subscription:v2:sgp_dotcom": nil},
		}

		removeUnseenTokens(seen, cache, log.Scoped("test"))
		assert.Equal(t, cache.calls, []struct{ call, key string }{{"ListAllKeys", ""}})
	})
}

type fakeListingCache struct {
	state map[string][]byte
	calls []struct{ call, key string }
}

var _ listingCache = &fakeListingCache{}

func (m *fakeListingCache) Set(key string, responseBytes []byte) {
	m.state[key] = responseBytes
	m.calls = append(m.calls, struct{ call, key string }{"Set", key})
}

func (m *fakeListingCache) ListAllKeys() []string {
	m.calls = append(m.calls, struct{ call, key string }{"ListAllKeys", ""})
	return maps.Keys(m.state)
}

func (m *fakeListingCache) Get(key string) ([]byte, bool) {
	m.calls = append(m.calls, struct{ call, key string }{"Get", key})
	if v, ok := m.state[key]; ok {
		return v, ok
	}
	return nil, false
}

func (m *fakeListingCache) Delete(key string) {
	m.calls = append(m.calls, struct{ call, key string }{"Delete", key})
	delete(m.state, key)
}
