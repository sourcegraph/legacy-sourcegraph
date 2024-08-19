package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func (r *UserResolver) AffiliatedNamespaces(ctx context.Context) (gqlutil.SliceConnectionResolver[*NamespaceResolver], error) {
	// Start with the user's own account (which is a namespace they are always affiliated with).
	namespaces := []*NamespaceResolver{
		{Namespace: r},
	}

	// 🚨 SECURITY: Only the user and admins are allowed to access the user's affiliated namespaces
	// because it reveals the user's org memberships, which are private.
	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}
	// Add all organizations the user is a member of.
	orgs, err := r.db.Orgs().GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		namespaces = append(namespaces, &NamespaceResolver{Namespace: &OrgResolver{db: r.db, org: org}})
	}

	return newNamespaceConnection(namespaces), nil
}

func (visitorResolver) AffiliatedNamespaces(context.Context) (gqlutil.SliceConnectionResolver[*NamespaceResolver], error) {
	return newNamespaceConnection(nil), nil
}
