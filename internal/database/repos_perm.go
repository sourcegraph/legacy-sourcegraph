package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var errPermissionsUserMappingConflict = errors.New("The permissions user mapping (site configuration `permissions.userMapping`) cannot be enabled when other authorization providers are in use, please contact site admin to resolve it.")

// ensure you use LOCAL to clear after tx
const ensureAuthzCondsFmt = `
	SET LOCAL ROLE sg_service;
	SET LOCAL rls.bypass = %v;
	SET LOCAL rls.user_id = %v;
	SET LOCAL rls.use_permissions_user_mapping = %v;
	SET LOCAL rls.permission = read;
`

func WithAuthzConds(ctx context.Context, db dbutil.DB) (dbutil.DB, func(error) error, error) {
	handle := basestore.NewHandleWithDB(db, sql.TxOptions{})
	tx, err := handle.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Copied from AuthzQueryConds below

	authzAllowByDefault, authzProviders := authz.GetProviders()
	usePermissionsUserMapping := globals.PermissionsUserMapping().Enabled

	// 🚨 SECURITY: Blocking access to all repositories if both code host authz
	// provider(s) and permissions user mapping are configured.
	if usePermissionsUserMapping {
		if len(authzProviders) > 0 {
			return nil, nil, errPermissionsUserMappingConflict
		}
		authzAllowByDefault = false
	}

	authenticatedUserID := int32(0)

	// Authz is bypassed when the request is coming from an internal actor or there
	// is no authz provider configured and access to all repositories are allowed by
	// default. Authz can be bypassed by site admins unless
	// conf.AuthEnforceForSiteAdmins is set to "true".
	bypassAuthz := isInternalActor(ctx) || (authzAllowByDefault && len(authzProviders) == 0)
	if !bypassAuthz && actor.FromContext(ctx).IsAuthenticated() {
		currentUser, err := Users(tx.DB()).GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, nil, tx.Done(err)
		}
		authenticatedUserID = currentUser.ID
		bypassAuthz = currentUser.SiteAdmin && !conf.Get().AuthzEnforceForSiteAdmins
	}

	// End of copy

	_, err = tx.DB().ExecContext(ctx, fmt.Sprintf(
		ensureAuthzCondsFmt,
		bypassAuthz,
		authenticatedUserID,
		usePermissionsUserMapping,
	))
	if err != nil {
		return nil, nil, tx.Done(err)
	}
	return tx.DB(), tx.Done, nil
}

// AuthzQueryConds returns a query clause for enforcing repository permissions.
// It uses `repo` as the table name to filter out repository IDs and should be
// used as an AND condition in a complete SQL query.
func AuthzQueryConds(ctx context.Context, db dbutil.DB) (*sqlf.Query, error) {
	authzAllowByDefault, authzProviders := authz.GetProviders()
	usePermissionsUserMapping := globals.PermissionsUserMapping().Enabled

	// 🚨 SECURITY: Blocking access to all repositories if both code host authz
	// provider(s) and permissions user mapping are configured.
	if usePermissionsUserMapping {
		if len(authzProviders) > 0 {
			return nil, errPermissionsUserMappingConflict
		}
		authzAllowByDefault = false
	}

	authenticatedUserID := int32(0)

	// Authz is bypassed when the request is coming from an internal actor or there
	// is no authz provider configured and access to all repositories are allowed by
	// default. Authz can be bypassed by site admins unless
	// conf.AuthEnforceForSiteAdmins is set to "true".
	bypassAuthz := isInternalActor(ctx) || (authzAllowByDefault && len(authzProviders) == 0)
	if !bypassAuthz && actor.FromContext(ctx).IsAuthenticated() {
		currentUser, err := Users(db).GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, err
		}
		authenticatedUserID = currentUser.ID
		bypassAuthz = currentUser.SiteAdmin && !conf.Get().AuthzEnforceForSiteAdmins
	}

	// TODO: if we're in dotcom mode, and we're authenticated, NEVER bypass authz.

	q := authzQuery(bypassAuthz,
		usePermissionsUserMapping,
		authenticatedUserID,
		authz.Read, // Note: We currently only support read for repository permissions.
	)
	return q, nil
}

func authzQuery(bypassAuthz, usePermissionsUserMapping bool, authenticatedUserID int32, perms authz.Perms) *sqlf.Query {
	const queryFmtString = `(
    %s                            -- TRUE or FALSE to indicate whether to bypass the check
OR  (
	NOT %s                        -- Disregard unrestricted state when permissions user mapping is enabled
	AND (
		NOT repo.private          -- Happy path of non-private repositories
		OR  EXISTS (              -- Each external service defines if repositories are unrestricted
			SELECT
			FROM external_services AS es
			JOIN external_service_repos AS esr ON (
					esr.external_service_id = es.id
				AND esr.repo_id = repo.id
				AND es.unrestricted = TRUE
				AND es.deleted_at IS NULL
			)
			LIMIT 1
		)
	)
)
OR EXISTS ( -- We assume that all repos added by the authenticated user should be shown
  SELECT 1
  FROM external_service_repos
  WHERE repo_id = repo.id
  AND user_id = %s
)
OR (                             -- Restricted repositories require checking permissions
	SELECT object_ids_ints @> INTSET(repo.id)
	FROM user_permissions
	WHERE
		user_id = %s
	AND permission = %s
	AND object_type = 'repos'
)
)
`

	return sqlf.Sprintf(queryFmtString,
		bypassAuthz,
		usePermissionsUserMapping,
		authenticatedUserID,
		authenticatedUserID,
		perms.String(),
	)
}

// isInternalActor returns true if the actor represents an internal agent (i.e., non-user-bound
// request that originates from within Sourcegraph itself).
//
// 🚨 SECURITY: internal requests bypass authz provider permissions checks, so correctness is
// important here.
func isInternalActor(ctx context.Context) bool {
	return actor.FromContext(ctx).Internal
}
