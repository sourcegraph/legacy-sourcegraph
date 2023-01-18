package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSiteConfiguration(t *testing.T) {
	t.Run("authenticated as non-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		_, err := newSchemaResolver(db, gitserver.NewClient(db)).Site().Configuration(ctx)

		if err == nil || !errors.Is(err, auth.ErrMustBeSiteAdmin) {
			t.Fatalf("err: want %q but got %v", auth.ErrMustBeSiteAdmin, err)
		}
	})
}

func TestSiteConfigurationHistory(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: stubs.users[0].ID})
	schemaResolver, err := newSchemaResolver(stubs.db, gitserver.NewClient(stubs.db)).Site().Configuration(ctx)
	if err != nil {
		t.Fatalf("failed to create schemaResolver: %v", err)
	}

	testCases := []struct {
		name                  string
		args                  *graphqlutil.ConnectionResolverArgs
		expectedSiteConfigIDs []int32
	}{
		{
			name: "first: 2",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(2),
			},
			expectedSiteConfigIDs: []int32{5, 4},
		},
		{
			name: "first: 5 (exact number of items that exist in the database)",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(5),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name: "first: 20 (more items than what exists in the database)",
			args: &graphqlutil.ConnectionResolverArgs{
				First: int32Ptr(20),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},

		{
			name: "last: 2",
			args: &graphqlutil.ConnectionResolverArgs{
				Last: int32Ptr(2),
			},
			expectedSiteConfigIDs: []int32{2, 1},
		},
		{
			name: "last: 5 (exact number of items that exist in the database)",
			args: &graphqlutil.ConnectionResolverArgs{
				Last: int32Ptr(5),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
		{
			name: "last: 20 (more items than what exists in the database)",
			args: &graphqlutil.ConnectionResolverArgs{
				Last: int32Ptr(5),
			},
			expectedSiteConfigIDs: []int32{5, 4, 3, 2, 1},
		},
	}

	for _, tc := range testCases {
		connectionResolver, err := schemaResolver.History(ctx, tc.args)
		if err != nil {
			t.Fatalf("failed to get history: %v", err)
		}

		siteConfigChangeResolvers, err := connectionResolver.Nodes(ctx)
		if err != nil {
			t.Fatalf("failed to get nodes: %v", err)
		}

		for i, resolver := range siteConfigChangeResolvers {
			if resolver.siteConfig.ID != tc.expectedSiteConfigIDs[i] {
				t.Errorf("position %d: expected siteConfig.ID %d, but got %d", i, tc.expectedSiteConfigIDs[i], resolver.siteConfig.ID)
			}
		}
	}

}
