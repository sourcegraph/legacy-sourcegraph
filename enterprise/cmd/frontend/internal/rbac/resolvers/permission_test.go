package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func TestPermissionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	user := createTestUser(t, db, false)
	admin := createTestUser(t, db, true)

	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))
	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))

	perm, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: "BATCHCHANGES",
		Action:    "READ",
	})
	if err != nil {
		t.Fatal(err)
	}

	s, err := newSchema(db, &Resolver{
		db:     db,
		logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}

	mpid := string(marshalPermissionID(perm.ID))

	t.Run("as non site-administrator", func(t *testing.T) {
		input := map[string]any{"permission": mpid}
		var response struct{ Node apitest.Permission }
		errs := apitest.Exec(userCtx, t, s, input, &response, queryPermissionNode)

		assert.Len(t, errs, 1)
		assert.Equal(t, errs[0].Message, "must be site admin")
	})

	t.Run(" as site-administrator", func(t *testing.T) {
		want := apitest.Permission{
			Typename:  "Permission",
			ID:        mpid,
			Namespace: perm.Namespace,
			Action:    perm.Action,
			CreatedAt: gqlutil.DateTime{Time: perm.CreatedAt.Truncate(time.Second)},
		}

		input := map[string]any{"permission": mpid}
		var response struct{ Node apitest.Permission }
		apitest.MustExec(adminCtx, t, s, input, &response, queryPermissionNode)
		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

const queryPermissionNode = `
query ($permission: ID!) {
	node(id: $permission) {
		__typename

		... on Permission {
			id
			namespace
			action
			createdAt
		}
	}
}
`
