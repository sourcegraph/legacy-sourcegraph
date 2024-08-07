package dotcomdb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/hexops/autogold/v2"
	pgxstdlibv4 "github.com/jackc/pgx/v4/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/codyaccessservice"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/databasetest"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/importer"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/dotcomproductsubscriptiontest"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	sgdatabase "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newTestDotcomReader(t *testing.T, opts dotcomdb.ReaderOptions) (sgdatabase.DB, *dotcomdb.Reader) {
	ctx := context.Background()

	// Set up a Sourcegraph test database.
	sgtestdb := dbtest.NewDB(t)

	// HACK: Extract the underlying pgx connection from the Sourcegraph test
	// database, so that we can tease out the connection string that was used
	// to connect to it, since we want to use pgx/v5 directly as offered by MSP
	// utilities. This is sneaky but easier than changing the extensive
	// infrastructure around dbtest.
	var connString string
	c, err := sgtestdb.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, c.Raw(func(driverConn any) error {
		// Unwrap all the way down. If something wraps the pgx conn in a
		// non-unwrappable fashion, that would be unfortunate, we'll end up
		// panicking in the next section.
		for {
			if wc, ok := driverConn.(dbconn.UnwrappableConn); ok {
				driverConn = wc.Raw()
			} else {
				break
			}
		}
		// Sourcegraph database uses pgxv4 under the hood, so we cast into the
		// v4 version of pgx internals
		connString = driverConn.(*pgxstdlibv4.Conn).Conn().Config().ConnString()
		return nil
	}))

	// Now create a new connection using the conn string 😎
	t.Logf("pgx.Connect %q", connString)
	conn, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Make sure it works!
	r := dotcomdb.NewReader(conn, opts)
	require.NoError(t, r.Ping(ctx))

	return sgdatabase.NewDB(logtest.Scoped(t), sgtestdb), r
}

type mockCodyAccessV1Store struct{ codyaccessservice.StoreV1 }

func (mockCodyAccessV1Store) IntrospectSAMSToken(context.Context, string) (*sams.IntrospectTokenResponse, error) {
	return &sams.IntrospectTokenResponse{
		Active:    true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scopes:    scopes.Scopes{scopes.ToScope(scopes.ServiceEnterprisePortal, scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionRead)},
	}, nil
}

func mockAuthenticatedServiceRequest[T any](message *T) *connect.Request[T] {
	req := connect.NewRequest(message)
	req.Header().Set("authorization", "bearer foobar")
	return req
}

type mockedData struct {
	targetSubscriptionID  string
	accessTokens          []string
	createdLicenses       int
	createdSubscriptions  int
	archivedSubscriptions int
}

func setupDBAndInsertMockLicense(t *testing.T, dotcomdb sgdatabase.DB, info license.Info, cgAccess *graphqlbackend.UpdateCodyGatewayAccessInput) mockedData {
	start := time.Now()

	ctx := context.Background()
	subscriptionsdb := dotcomproductsubscriptiontest.NewSubscriptionsDB(t, dotcomdb)
	licensesdb := dotcomproductsubscriptiontest.NewLicensesDB(t, dotcomdb)
	result := mockedData{}

	{
		// Create a different subscription and license that's rubbish,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "barbaz"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-barbaz", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{licensing.DevTag},
		})
		require.NoError(t, err)
		result.createdLicenses += 1
	}

	{
		// Create a different subscription and license that's archived,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "archived"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-archived", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{licensing.DevTag},
		})
		require.NoError(t, err)
		result.createdLicenses += 1
		// Archive the subscription
		require.NoError(t, subscriptionsdb.Archive(ctx, sub))
		result.archivedSubscriptions += 1
	}

	{
		// Create a different subscription and license that's not a dev tag,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "not-devlicense"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-not-devlicense", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{},
		})
		require.NoError(t, err)
	}

	// Create the subscription we will assert against
	u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "user"})
	require.NoError(t, err)
	subid, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
	require.NoError(t, err)
	result.createdSubscriptions += 1
	result.targetSubscriptionID = subid
	// Insert a rubbish license first, CreatedAt is not used (creation time is
	// inferred from insert time) so we need to do this first
	key1 := t.Name() + "-1"
	result.accessTokens = append(result.accessTokens, license.GenerateLicenseKeyBasedAccessToken(key1))
	_, err = licensesdb.Create(ctx, subid, key1, 2, license.Info{
		CreatedAt: info.CreatedAt.Add(-time.Hour),
		ExpiresAt: info.ExpiresAt.Add(-time.Minute), // should expire first, but not be expired
		Tags:      []string{licensing.DevTag},
	})
	require.NoError(t, err)
	result.createdLicenses += 1
	// Now create the actual requested license, and test that this is the one
	// that gets used.
	key2 := t.Name() + "-2"
	result.accessTokens = append(result.accessTokens, license.GenerateLicenseKeyBasedAccessToken(key2))
	_, err = licensesdb.Create(ctx, subid, key2, 2, info)
	require.NoError(t, err)
	result.createdLicenses += 1
	// Configure Cody Gateway access for this subscription as we do today
	if cgAccess != nil {
		require.NoError(t, subscriptionsdb.UpdateCodyGatewayAccess(ctx, subid, *cgAccess))
	}

	{
		// Create another different subscription and license that's also rubbish,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "foobar"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-foobar", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{licensing.DevTag},
		})
		require.NoError(t, err)
		result.createdLicenses += 1
	}

	t.Logf("Setup complete in %s", time.Since(start).String())
	return result
}

func mustWithCtx[T any](t *testing.T, fn func(context.Context) (T, error)) T {
	v, err := fn(context.Background())
	require.NoError(t, err)
	return v
}

func TestGetCodyGatewayAccessAttributes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	for i, tc := range []struct {
		name     string
		info     license.Info
		cgAccess graphqlbackend.UpdateCodyGatewayAccessInput
	}{{
		name: "disabled",
		info: license.Info{
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 10,
			Tags:      []string{licensing.DevTag},
		},
		cgAccess: graphqlbackend.UpdateCodyGatewayAccessInput{
			Enabled: pointers.Ptr(false),
		},
	}, {
		name: "user and tags",
		info: license.Info{
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 321,
			Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
		},
		cgAccess: graphqlbackend.UpdateCodyGatewayAccessInput{Enabled: pointers.Ptr(true)},
	}, {
		name: "overrides",
		info: license.Info{
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 10,
			Tags:      []string{licensing.DevTag},
		},
		cgAccess: graphqlbackend.UpdateCodyGatewayAccessInput{
			Enabled:                                 pointers.Ptr(true),
			ChatCompletionsRateLimit:                pointers.Ptr[graphqlbackend.BigInt](123),
			CodeCompletionsRateLimitIntervalSeconds: pointers.Ptr[int32](456),
			EmbeddingsRateLimit:                     pointers.Ptr[graphqlbackend.BigInt](789),
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			dotcomdb, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
				DevOnly: true,
			})

			// First, set up a subscription and license and some other rubbish
			// data to ensure we only get the license we want.
			mock := setupDBAndInsertMockLicense(t, dotcomdb, tc.info, &tc.cgAccess)

			// Now import the data for parity so we can compare against the
			// Enterprise Portal implementation
			epDB := &database.DB{
				DB: databasetest.NewTestDB(t, "ep-dotcomdb", fmt.Sprintf("get-attributes-%d", i), databasetest.Tables(t)...),
			}
			err := importer.NewHandler(ctx, logtest.Scoped(t), dotcomreader, epDB, nil).
				ImportSubscriptions(ctx)
			require.NoError(t, err)
			codyAccessService := codyaccessservice.NewHandlerV1(logtest.Scoped(t), mockCodyAccessV1Store{
				StoreV1: codyaccessservice.NewStoreV1(codyaccessservice.StoreV1Options{
					DB: epDB,
				}),
			})

			t.Run("by subscription ID", func(t *testing.T) {
				t.Parallel()

				attr, err := dotcomreader.GetCodyGatewayAccessAttributesBySubscription(ctx, mock.targetSubscriptionID)
				require.NoError(t, err)
				access, err := codyAccessService.GetCodyGatewayAccess(ctx, mockAuthenticatedServiceRequest(&codyaccessv1.GetCodyGatewayAccessRequest{
					Query: &codyaccessv1.GetCodyGatewayAccessRequest_SubscriptionId{
						SubscriptionId: mock.targetSubscriptionID,
					},
				}))
				require.NoError(t, err)

				validateAccessAttributes(t, dotcomdb, mock, attr, access.Msg.GetAccess(), tc.info)
			})

			t.Run("by access token", func(t *testing.T) {
				t.Parallel()

				for i, token := range mock.accessTokens {
					t.Run(fmt.Sprintf("token %d", i), func(t *testing.T) {
						token := token
						t.Parallel()

						attr, err := dotcomreader.GetCodyGatewayAccessAttributesByAccessToken(ctx, token)
						require.NoError(t, err)
						access, err := codyAccessService.GetCodyGatewayAccess(ctx, mockAuthenticatedServiceRequest(&codyaccessv1.GetCodyGatewayAccessRequest{
							Query: &codyaccessv1.GetCodyGatewayAccessRequest_AccessToken{
								AccessToken: token,
							},
						}))
						require.NoError(t, err)
						validateAccessAttributes(t, dotcomdb, mock, attr, access.Msg.GetAccess(), tc.info)

						t.Run("compare with dotcom tokens DB", func(t *testing.T) {
							subID, err := dotcomproductsubscriptiontest.NewTokensDB(t, dotcomdb).
								LookupProductSubscriptionIDByAccessToken(ctx, token)
							require.NoError(t, err)
							assert.Equal(t, subID, attr.SubscriptionID)
						})
					})
				}
			})
		})
	}
}

func validateAccessAttributes(t *testing.T, dotcomdb sgdatabase.DB, mock mockedData, attr *dotcomdb.CodyGatewayAccessAttributes, access *codyaccessv1.CodyGatewayAccess, info license.Info) {
	assert.Equal(t, mock.targetSubscriptionID, attr.SubscriptionID)
	assert.Equal(t, subscriptionsv1.EnterpriseSubscriptionIDPrefix+mock.targetSubscriptionID, access.SubscriptionId)

	assert.Equal(t, int(info.UserCount), *attr.ActiveLicenseUserCount)
	assert.Len(t, attr.LicenseKeyHashes, 2)

	assert.Equal(t, mock.accessTokens, attr.GenerateAccessTokens())
	var protoAccessTokens []string
	for _, t := range access.AccessTokens {
		protoAccessTokens = append(protoAccessTokens, t.GetToken())
	}
	assert.ElementsMatch(t, mock.accessTokens, protoAccessTokens)

	limits := attr.EvaluateRateLimits()

	// Validate against the expected values as produced by existing resolvers
	expected := dotcomproductsubscriptiontest.NewCodyGatewayAccessResolver(t, dotcomdb, mock.targetSubscriptionID)
	assert.Equal(t, expected.Enabled(), attr.CodyGatewayEnabled)
	if !expected.Enabled() {
		// We don't care about the rest of the attributes if access is disabled,
		// the resolver returns nil for everything
		return
	}
	for _, compare := range []struct {
		name     string
		expected graphqlbackend.CodyGatewayRateLimit
		got      licensing.CodyGatewayRateLimit
		gotProto *codyaccessv1.CodyGatewayRateLimit
	}{{
		name:     "Chat",
		expected: mustWithCtx(t, expected.ChatCompletionsRateLimit),
		got:      limits.Chat,
		gotProto: access.ChatCompletionsRateLimit,
	}, {
		name:     "Code",
		expected: mustWithCtx(t, expected.CodeCompletionsRateLimit),
		got:      limits.Code,
		gotProto: access.CodeCompletionsRateLimit,
	}, {
		name:     "Embeddings",
		expected: mustWithCtx(t, expected.EmbeddingsRateLimit),
		got:      limits.Embeddings,
		gotProto: access.EmbeddingsRateLimit,
	}} {
		t.Run(compare.name, func(t *testing.T) {
			// We only care about limit and interval now
			assert.Equal(t, int64(compare.expected.Limit()), compare.got.Limit, "Limit")
			assert.Equal(t, compare.expected.IntervalSeconds(), compare.got.IntervalSeconds, "IntervalSeconds")

			assert.Equal(t, uint64(compare.expected.Limit()), compare.gotProto.Limit, "Limit")
			assert.Equal(t, int64(compare.expected.IntervalSeconds()), compare.gotProto.IntervalDuration.Seconds, "IntervalSeconds")
		})
	}
}

func TestGetAllCodyGatewayAccessAttributes(t *testing.T) {
	t.Parallel()
	dotcomDB, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
		DevOnly: true,
	})

	info := license.Info{
		CreatedAt: time.Now().Add(-30 * time.Minute),
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UserCount: 321,
		Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
	}
	cgAccess := graphqlbackend.UpdateCodyGatewayAccessInput{Enabled: pointers.Ptr(true)}
	mock := setupDBAndInsertMockLicense(t, dotcomDB, info, &cgAccess)

	// Now import the data for parity so we can compare against the
	// Enterprise Portal implementation
	ctx := context.Background()
	epDB := &database.DB{
		DB: databasetest.NewTestDB(t, "ep-dotcomdb", "get-all-attributes", databasetest.Tables(t)...),
	}
	err := importer.NewHandler(ctx, logtest.Scoped(t), dotcomreader, epDB, nil).
		ImportSubscriptions(ctx)
	require.NoError(t, err)
	codyAccessService := codyaccessservice.NewHandlerV1(logtest.Scoped(t), mockCodyAccessV1Store{
		StoreV1: codyaccessservice.NewStoreV1(codyaccessservice.StoreV1Options{
			DB: epDB,
		}),
	})

	attrs, err := dotcomreader.GetAllCodyGatewayAccessAttributes(ctx)
	require.NoError(t, err)
	assert.Len(t, attrs, 3) // 3 subscriptions created in setupDBAndInsertMockLicense

	accesses, err := codyAccessService.ListCodyGatewayAccesses(ctx, mockAuthenticatedServiceRequest(&codyaccessv1.ListCodyGatewayAccessesRequest{}))
	require.NoError(t, err)
	assert.Len(t, accesses.Msg.GetAccesses(), 3) // 3 subscriptions created in setupDBAndInsertMockLicense

	var (
		foundAttr   *dotcomdb.CodyGatewayAccessAttributes
		foundAccess *codyaccessv1.CodyGatewayAccess
	)
	for _, attr := range attrs {
		if attr.SubscriptionID == mock.targetSubscriptionID {
			foundAttr = attr
		} else {
			assert.False(t, attr.CodyGatewayEnabled)
		}
	}
	for _, access := range accesses.Msg.GetAccesses() {
		if access.SubscriptionId == subscriptionsv1.EnterpriseSubscriptionIDPrefix+mock.targetSubscriptionID {
			foundAccess = access
		} else {
			assert.False(t, access.Enabled)
		}
	}
	require.NotNil(t, foundAttr)
	require.NotNil(t, foundAccess)

	validateAccessAttributes(t, dotcomDB, mock, foundAttr, foundAccess, info)

}

func TestListEnterpriseSubscriptionLicenses(t *testing.T) {
	t.Parallel()

	db, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
		DevOnly: true,
	})
	info := license.Info{
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UserCount: 321,
		Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
	}
	rootTestName := t.Name()
	mock := setupDBAndInsertMockLicense(t, db, info, nil)

	assertMockLicense := func(t *testing.T, l *dotcomdb.LicenseAttributes) {
		assert.Equal(t, mock.targetSubscriptionID, l.SubscriptionID)
		assert.Equal(t, info.UserCount, uint(*l.UserCount))
		assert.NotEmpty(t, l.CreatedAt)
		assert.Equal(t,
			info.ExpiresAt.Format(time.DateTime),
			l.ExpiresAt.Format(time.DateTime))
		assert.Equal(t, info.Tags, l.Tags)
	}

	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		filters  []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter
		pageSize int
		expect   func(t *testing.T, licenses []*dotcomdb.LicenseAttributes)
	}{{
		name:    "no filters",
		filters: nil,
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			// Only unarchived subscriptions
			assert.Len(t, licenses, mock.createdLicenses-mock.archivedSubscriptions)
		},
	}, {
		name: "filter by subscription ID",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
				SubscriptionId: mock.targetSubscriptionID,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, 2) // setupDBAndInsertMockLicense adds 2
			// The first one is our most recently created one, i.e. the one we
			// requested in setupDBAndInsertMockLicense
			assertMockLicense(t, licenses[0])
		},
	}, {
		name: "filter by subscription ID and limit 1",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
				SubscriptionId: mock.targetSubscriptionID,
			},
		}},
		pageSize: 1,
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, 1)
			assertMockLicense(t, licenses[0])
		},
	}, {
		name: "filter by subscription ID and not archived",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
				SubscriptionId: mock.targetSubscriptionID,
			},
		}, {
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
				IsRevoked: false,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			// setupDBAndInsertMockLicense adds 2 for the target subscription,
			// both unarchived
			assert.Len(t, licenses, 2)
			for _, l := range licenses {
				assert.Equal(t, mock.targetSubscriptionID, l.SubscriptionID)
			}
		},
	}, {
		name: "filter by is archived",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
				IsRevoked: true,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, mock.archivedSubscriptions)
		},
	}, {
		name: "filter by not archived",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
				IsRevoked: false,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, mock.createdLicenses-mock.archivedSubscriptions)
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			licenses, err := dotcomreader.ListEnterpriseSubscriptionLicenses(ctx, tc.filters, tc.pageSize)
			require.NoError(t, err)
			for _, l := range licenses {
				// Each mock license key contains a variation of the root test name
				assert.Contains(t, l.LicenseKey, rootTestName)
			}
			if tc.expect != nil {
				tc.expect(t, licenses)
			}
		})
	}
}

func TestListEnterpriseSubscriptions(t *testing.T) {
	t.Run("devonly", func(t *testing.T) {
		t.Parallel()

		db, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
			DevOnly: true,
		})
		info := license.Info{
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 321,
			Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
		}
		mock := setupDBAndInsertMockLicense(t, db, info, nil)

		ss, err := dotcomreader.ListEnterpriseSubscriptions(
			context.Background(),
			dotcomdb.ListEnterpriseSubscriptionsOptions{})
		require.NoError(t, err)
		// We expect 1 less subscription because one of the subscriptions does not
		// have a dev/internal license
		assert.Len(t, ss, mock.createdSubscriptions-mock.archivedSubscriptions-1)
		for _, s := range ss {
			s.CreatedAt = time.Time{} // zero time for autogold
		}
		autogold.Expect("foobar - 0001-01-01 00:00:00").Equal(t, ss[0].GenerateDisplayName())
		autogold.Expect("user - 0001-01-01 00:00:00").Equal(t, ss[1].GenerateDisplayName())
		autogold.Expect("barbaz - 0001-01-01 00:00:00").Equal(t, ss[2].GenerateDisplayName())

		var found bool
		for _, s := range ss {
			if s.ID == mock.targetSubscriptionID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("not devonly", func(t *testing.T) {
		t.Parallel()

		db, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
			DevOnly: false,
		})
		info := license.Info{
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 321,
			Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
		}
		mock := setupDBAndInsertMockLicense(t, db, info, nil)

		ss, err := dotcomreader.ListEnterpriseSubscriptions(
			context.Background(),
			dotcomdb.ListEnterpriseSubscriptionsOptions{})
		require.NoError(t, err)
		// all subscriptions included, minus the ones without any license
		assert.Len(t, ss, mock.createdSubscriptions-1)
	})
}
