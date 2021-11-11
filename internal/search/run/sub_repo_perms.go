package run

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// applySubRepoPerms drops matches the actor in the given context does not have read access to.
func applySubRepoPerms(ctx context.Context, srp authz.SubRepoPermissionChecker, event *streaming.SearchEvent) error {
	actor := actor.FromContext(ctx)
	errs := &multierror.Error{}
	authorized := event.Results[:0]
	for _, match := range event.Results {
		key := match.Key()
		perms, err := authz.ActorPermissions(ctx, srp, actor, authz.RepoContent{
			Repo: key.Repo,
			Path: key.Path,
		})
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("applySubRepoPerms: failed to check sub-repo permissions (actor.uid: %d, match.key: %v): %w",
				actor.UID, key, err))
		}
		if perms.Include(authz.Read) {
			authorized = append(authorized, match)
		}
	}

	// Only keep authorized matches
	event.Results = authorized

	return errs.ErrorOrNil()
}
