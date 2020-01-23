package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

var now = time.Now().Truncate(time.Microsecond).UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now)).Truncate(time.Microsecond)
}

func mustParseGraphQLSchema(t *testing.T, db *sql.DB) *graphql.Schema {
	t.Helper()

	schema, err := graphqlbackend.NewSchema(nil, nil, NewResolver(db, clock))
	if err != nil {
		t.Fatal(err)
	}

	return schema
}

func TestResolver_SetRepositoryPermissionsForUsers(t *testing.T) {
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.Repos.Get = func(context.Context, api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: 1}, nil
	}
	defer func() {
		db.Mocks.Users.GetByCurrentAuthUser = nil
		db.Mocks.Repos.Get = nil
	}()

	tests := []struct {
		name                 string
		config               *schema.PermissionsUserMapping
		mockVerifiedEmails   []*db.UserEmail
		mockUsers            []*types.User
		gqlTests             []*gqltesting.Test
		expectUserIDs        []uint32
		expectPendingBindIDs []string
	}{
		{
			name: "set permissions via email",
			config: &schema.PermissionsUserMapping{
				BindID: "email",
			},
			mockVerifiedEmails: []*db.UserEmail{
				{
					UserID: 1,
					Email:  "alice@example.com",
				},
			},
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				mutation {
					setRepositoryPermissionsForUsers(
						repository: "UmVwb3NpdG9yeToxMzA5Mjg1",
						bindIDs: ["alice@example.com", "bob"]) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"setRepositoryPermissionsForUsers": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			expectUserIDs:        []uint32{1},
			expectPendingBindIDs: []string{"bob"},
		},
		{
			name: "set permissions via username",
			config: &schema.PermissionsUserMapping{
				BindID: "username",
			},
			mockUsers: []*types.User{
				{
					ID:       1,
					Username: "alice",
				},
			},
			gqlTests: []*gqltesting.Test{
				{
					Schema: mustParseGraphQLSchema(t, nil),
					Query: `
				mutation {
					setRepositoryPermissionsForUsers(
						repository: "UmVwb3NpdG9yeToxMzA5Mjg1",
						bindIDs: ["alice", "bob"]) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"setRepositoryPermissionsForUsers": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			expectUserIDs:        []uint32{1},
			expectPendingBindIDs: []string{"bob"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals.SetPermissionsUserMapping(test.config)

			db.Mocks.UserEmails.GetVerifiedEmails = func(_ context.Context, emails ...string) ([]*db.UserEmail, error) {
				return test.mockVerifiedEmails, nil
			}
			db.Mocks.Users.GetByUsernames = func(_ context.Context, usernames ...string) ([]*types.User, error) {
				return test.mockUsers, nil
			}
			edb.Mocks.Perms.SetRepoPermissions = func(_ context.Context, p *iauthz.RepoPermissions) error {
				ids := p.UserIDs.ToArray()
				if diff := cmp.Diff(test.expectUserIDs, ids); diff != "" {
					return fmt.Errorf("p.UserIDs: %v", diff)
				}
				return nil
			}
			edb.Mocks.Perms.SetRepoPendingPermissions = func(_ context.Context, bindIDs []string, _ *iauthz.RepoPermissions) error {
				if diff := cmp.Diff(test.expectPendingBindIDs, bindIDs); diff != "" {
					return fmt.Errorf("bindIDs: %v", diff)
				}
				return nil
			}
			defer func() {
				db.Mocks.UserEmails.GetVerifiedEmails = nil
				db.Mocks.Users.GetByUsernames = nil
				edb.Mocks.Perms.SetRepoPermissions = nil
				edb.Mocks.Perms.SetRepoPendingPermissions = nil
			}()

			gqltesting.RunTests(t, test.gqlTests)
		})
	}
}
