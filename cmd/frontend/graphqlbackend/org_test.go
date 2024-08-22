package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestOrganization(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, &database.ErrOrgMemberNotFound{})

	orgs := dbmocks.NewMockOrgStore()
	mockedOrg := types.Org{ID: 1, Name: "acme"}
	orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)
	orgs.GetByIDFunc.SetDefaultReturn(&mockedOrg, nil)

	securityLogEvents := dbmocks.NewMockSecurityEventLogsStore()
	securityLogEvents.LogSecurityEventFunc.SetDefaultReturn(nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)
	db.SecurityEventLogsFunc.SetDefaultReturn(securityLogEvents)

	t.Run("can access organizations", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
				{
					organization(name: "acme") {
						name
					}
				}
			`,
				ExpectedResult: `
				{
					"organization": {
						"name": "acme"
					}
				}
			`,
			},
		})
	})
	mockrequire.Called(t, securityLogEvents.LogSecurityEventFunc)

}

func TestOrganizationMembers(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.ListByOrgFunc.SetDefaultReturn([]*types.User{
		{ID: 1, Username: "alice"},
		{ID: 2, Username: "bob"},
	}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultHook(func(ctx context.Context, orgID, userID int32) (*types.OrgMembership, error) {
		if orgID == 1 && userID == 1 {
			return &types.OrgMembership{OrgID: 1, UserID: 1}, nil
		}
		return nil, &database.ErrOrgMemberNotFound{}
	})

	orgs := dbmocks.NewMockOrgStore()
	mockedOrg := types.Org{ID: 1, Name: "acme"}
	orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)

	securityLogEvents := dbmocks.NewMockSecurityEventLogsStore()
	securityLogEvents.LogSecurityEventFunc.SetDefaultReturn(nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)
	db.SecurityEventLogsFunc.SetDefaultReturn(securityLogEvents)

	t.Run("org members can list members", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{Username: "alice", ID: 1}, nil)
		for _, isDotcom := range []bool{true, false} {
			t.Run(fmt.Sprintf("dotcom=%v", isDotcom), func(t *testing.T) {
				dotcom.MockSourcegraphDotComMode(t, isDotcom)
				RunTests(t, []*Test{
					{
						Schema: mustParseGraphQLSchema(t, db),
						Query: `
					{
						organization(name: "acme") {
							members {
								nodes { username }
							}
						}
					}
				`,
						ExpectedResult: `
					{
						"organization": {
							"members": {
								"nodes": [{"username": "alice"}, {"username": "bob"}]
							}
						}
					}
				`,
					},
				})
			})
		}
		mockrequire.Called(t, securityLogEvents.LogSecurityEventFunc)
	})

	t.Run("non-members", func(t *testing.T) {
		t.Run("can list members on non-dotcom", func(t *testing.T) {
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{Username: "xavier", ID: 10}, nil)
			dotcom.MockSourcegraphDotComMode(t, false)
			RunTests(t, []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
						{
							organization(name: "acme") {
								members {
									nodes { username }
								}
							}
						}
					`,
					ExpectedResult: `
						{
							"organization": {
								"members": {
									"nodes": [{"username": "alice"}, {"username": "bob"}]
								}
							}
						}
					`,
				},
			})
		})

		t.Run("who are site admin can list members on non-dotcom", func(t *testing.T) {
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{Username: "xavier", ID: 10, SiteAdmin: true}, nil)
			dotcom.MockSourcegraphDotComMode(t, true)
			RunTests(t, []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
						{
							organization(name: "acme") {
								members {
									nodes { username }
								}
							}
						}
					`,
					ExpectedResult: `
						{
							"organization": {
								"members": {
									"nodes": [{"username": "alice"}, {"username": "bob"}]
								}
							}
						}
					`,
				},
			})
		})

		t.Run("who are not site admin cannot list members on dotcom", func(t *testing.T) {
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{Username: "xavier", ID: 10}, nil)
			dotcom.MockSourcegraphDotComMode(t, true)
			RunTests(t, []*Test{
				{
					Schema: mustParseGraphQLSchema(t, db),
					Query: `
					{
						organization(name: "acme") {
							members {
								nodes { username }
							}
						}
					}
				`,
					ExpectedResult: `
				{
					"organization": null
				}
				`,
					ExpectedErrors: []*gqlerrors.QueryError{
						{
							Message: "current user is not an org member",
							Path:    []any{"organization", "members"},
						},
					},
				},
			})
		})
		mockrequire.Called(t, securityLogEvents.LogSecurityEventFunc)

	})
}

func TestCreateOrganization(t *testing.T) {
	userID := int32(1)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID, SiteAdmin: false}, nil)

	mockedOrg := types.Org{ID: 42, Name: "acme"}
	orgs := dbmocks.NewMockOrgStore()
	orgs.CreateFunc.SetDefaultReturn(&mockedOrg, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.CreateFunc.SetDefaultReturn(&types.OrgMembership{OrgID: mockedOrg.ID, UserID: userID}, nil)

	securityLogEvents := dbmocks.NewMockSecurityEventLogsStore()
	securityLogEvents.LogSecurityEventFunc.SetDefaultReturn(nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)
	db.SecurityEventLogsFunc.SetDefaultReturn(securityLogEvents)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: userID})

	t.Run("Creates organization", func(t *testing.T) {
		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `mutation CreateOrganization($name: String!, $displayName: String) {
				createOrganization(name: $name, displayName: $displayName) {
					id
                    name
				}
			}`,
			ExpectedResult: fmt.Sprintf(`
			{
				"createOrganization": {
					"id": "%s",
					"name": "%s"
				}
			}
			`, MarshalOrgID(mockedOrg.ID), mockedOrg.Name),
			Variables: map[string]any{
				"name": "acme",
			},
		})
		mockrequire.Called(t, securityLogEvents.LogSecurityEventFunc)
	})

	t.Run("Fails for unauthenticated user", func(t *testing.T) {
		dotcom.MockSourcegraphDotComMode(t, true)

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: context.Background(),
			Query: `mutation CreateOrganization($name: String!, $displayName: String) {
				createOrganization(name: $name, displayName: $displayName) {
					id
                    name
				}
			}`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Message: "no current user",
					Path:    []any{"createOrganization"},
				},
			},
			Variables: map[string]any{
				"name": "test",
			},
		})
	})

	t.Run("Fails for suspicious organization name", func(t *testing.T) {
		dotcom.MockSourcegraphDotComMode(t, true)

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `mutation CreateOrganization($name: String!, $displayName: String) {
				createOrganization(name: $name, displayName: $displayName) {
					id
                    name
				}
			}`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Message: `rejected suspicious name "test"`,
					Path:    []any{"createOrganization"},
				},
			},
			Variables: map[string]any{
				"name": "test",
			},
		})
	})
}

func TestAddOrganizationMember(t *testing.T) {
	userID := int32(2)
	userName := "add-org-member"
	orgID := int32(1)
	orgIDString := string(MarshalOrgID(orgID))

	orgs := dbmocks.NewMockOrgStore()
	orgs.GetByNameFunc.SetDefaultReturn(&types.Org{ID: orgID, Name: "acme"}, nil)

	users := dbmocks.NewMockUserStore()
	users.GetByUsernameFunc.SetDefaultReturn(&types.User{ID: 2, Username: userName}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, &database.ErrOrgMemberNotFound{})
	orgMembers.CreateFunc.SetDefaultReturn(&types.OrgMembership{OrgID: orgID, UserID: userID}, nil)

	featureFlags := dbmocks.NewMockFeatureFlagStore()
	featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

	// tests below depend on config being there
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}},
		EmailSmtp:     &schema.SMTPServerConfig{},
	}})

	// mock permission sync scheduling
	permssync.MockSchedulePermsSync = func(_ context.Context, logger log.Logger, _ database.DB, _ permssync.ScheduleSyncOpts) {}
	defer func() { permssync.MockSchedulePermsSync = nil }()

	securityLogEvents := dbmocks.NewMockSecurityEventLogsStore()
	securityLogEvents.LogSecurityEventFunc.SetDefaultReturn(nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)
	db.SecurityEventLogsFunc.SetDefaultReturn(securityLogEvents)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	t.Run("site admin is permitted", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `mutation AddUserToOrganization($organization: ID!, $username: String!) {
				addUserToOrganization(organization: $organization, username: $username) {
					alwaysNil
				}
			}`,
			ExpectedResult: `{
				"addUserToOrganization": {
					"alwaysNil": null
				}
			}`,
			Variables: map[string]any{
				"organization": orgIDString,
				"username":     userName,
			},
		})
	})

	t.Run("non-site admins are not permitted", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)
		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query: `mutation AddUserToOrganization($organization: ID!, $username: String!) {
				addUserToOrganization(organization: $organization, username: $username) {
					alwaysNil
				}
			}`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{{
				Message: "must be site admin",
				Path:    []any{"addUserToOrganization"},
			}},
			Variables: map[string]any{
				"organization": orgIDString,
				"username":     userName,
			},
		})
	})
}

func TestOrganizationRepositories_OSS(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					organization(name: "acme") {
						name,
						repositories {
							nodes {
								name
							}
						}
					}
				}
			`,
			ExpectedErrors: []*gqlerrors.QueryError{{
				Message:   `Cannot query field "repositories" on type "Org".`,
				Locations: []gqlerrors.Location{{Line: 5, Column: 7}},
				Rule:      "FieldsOnCorrectType",
			}},
			Context: ctx,
		},
	})
}

func TestNode_Org(t *testing.T) {
	orgs := dbmocks.NewMockOrgStore()
	orgs.GetByIDFunc.SetDefaultReturn(&types.Org{ID: 1, Name: "acme"}, nil)

	db := dbmocks.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					node(id: "T3JnOjE=") {
						id
						... on Org {
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "T3JnOjE=",
						"name": "acme"
					}
				}
			`,
		},
	})
}

func TestUnmarshalOrgID(t *testing.T) {
	t.Run("Valid org ID is parsed correctly", func(t *testing.T) {
		const id = int32(1)
		namespaceOrgID := relay.MarshalID("Org", id)
		orgID, err := UnmarshalOrgID(namespaceOrgID)
		assert.NoError(t, err)
		assert.Equal(t, id, orgID)
	})

	t.Run("Returns error for invalid org ID", func(t *testing.T) {
		const id = 1
		namespaceOrgID := relay.MarshalID("User", id)
		_, err := UnmarshalOrgID(namespaceOrgID)
		assert.Error(t, err)
	})
}

func TestMembersConnectionStore(t *testing.T) {
	ctx := context.Background()

	db := database.NewDB(logtest.Scoped(t), dbtest.NewDB(t))

	org, err := db.Orgs().Create(ctx, "test-org", nil)
	require.NoError(t, err)

	for i := range 10 {
		user, err := db.Users().Create(ctx, database.NewUser{
			Username:        "test" + strconv.Itoa(i),
			Email:           fmt.Sprintf("test%d@sourcegraph.com", i),
			EmailIsVerified: true,
		})
		require.NoError(t, err)
		_, err = db.OrgMembers().Create(ctx, org.ID, user.ID)
		require.NoError(t, err)
	}

	connectionStore := &membersConnectionStore{
		db:    db,
		orgID: org.ID,
	}

	gqlutil.TestConnectionResolverStoreSuite(t, connectionStore, nil)
}
