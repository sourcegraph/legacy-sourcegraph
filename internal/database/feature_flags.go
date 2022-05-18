package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	ff "github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type FeatureFlagStore interface {
	basestore.ShareableStore
	With(basestore.ShareableStore) FeatureFlagStore
	Transact(context.Context) (FeatureFlagStore, error)
	CreateFeatureFlag(context.Context, *ff.FeatureFlag) (*ff.FeatureFlag, error)
	UpdateFeatureFlag(context.Context, *ff.FeatureFlag) (*ff.FeatureFlag, error)
	DeleteFeatureFlag(context.Context, string) error
	CreateRollout(ctx context.Context, name string, rollout int32) (*ff.FeatureFlag, error)
	CreateBool(ctx context.Context, name string, value bool) (*ff.FeatureFlag, error)
	GetFeatureFlag(ctx context.Context, flagName string) (*ff.FeatureFlag, error)
	GetFeatureFlags(context.Context) ([]*ff.FeatureFlag, error)
	CreateOverride(context.Context, *ff.Override) (*ff.Override, error)
	DeleteOverride(ctx context.Context, orgID, userID *int32, flagName string) error
	UpdateOverride(ctx context.Context, orgID, userID *int32, flagName string, newValue bool) (*ff.Override, error)
	GetOverridesForFlag(context.Context, string) ([]*ff.Override, error)
	GetUserOverrides(context.Context, int32) ([]*ff.Override, error)
	GetOrgOverridesForUser(ctx context.Context, userID int32) ([]*ff.Override, error)
	GetOrgOverrideForFlag(ctx context.Context, orgID int32, flagName string) (*ff.Override, error)
	GetUserFlags(context.Context, int32) (map[string]bool, error)
	GetAnonymousUserFlags(ctx context.Context, anonymousUID string) (map[string]bool, error)
	GetGlobalFeatureFlags(context.Context) (map[string]bool, error)
	GetUserFlag(ctx context.Context, userID int32, flagName string) (*bool, error)
	GetAnonymousUserFlag(ctx context.Context, anonymousUID string, flagName string) (*bool, error)
	GetGlobalFeatureFlag(ctx context.Context, flagName string) (*bool, error)
	GetOrgFeatureFlag(ctx context.Context, orgID int32, flagName string) (bool, error)
}

type featureFlagStore struct {
	*basestore.Store
}

func FeatureFlagsWith(other basestore.ShareableStore) FeatureFlagStore {
	return &featureFlagStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (f *featureFlagStore) With(other basestore.ShareableStore) FeatureFlagStore {
	return &featureFlagStore{Store: f.Store.With(other)}
}

func (f *featureFlagStore) Transact(ctx context.Context) (FeatureFlagStore, error) {
	txBase, err := f.Store.Transact(ctx)
	return &featureFlagStore{Store: txBase}, err
}

func (f *featureFlagStore) CreateFeatureFlag(ctx context.Context, flag *ff.FeatureFlag) (*ff.FeatureFlag, error) {
	const newFeatureFlagFmtStr = `
		INSERT INTO feature_flags (
			flag_name,
			flag_type,
			bool_value,
			rollout
		) VALUES (
			%s,
			%s,
			%s,
			%s
		) RETURNING
			flag_name,
			flag_type,
			bool_value,
			rollout,
			created_at,
			updated_at,
			deleted_at
		;
	`
	var (
		flagType string
		boolVal  *bool
		rollout  *int32
	)
	switch {
	case flag.Bool != nil:
		flagType = "bool"
		boolVal = &flag.Bool.Value
	case flag.Rollout != nil:
		flagType = "rollout"
		rollout = &flag.Rollout.Rollout
	default:
		return nil, errors.New("feature flag must have exactly one type")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFeatureFlagFmtStr,
		flag.Name,
		flagType,
		boolVal,
		rollout))
	return scanFeatureFlag(row)
}

func (f *featureFlagStore) UpdateFeatureFlag(ctx context.Context, flag *ff.FeatureFlag) (*ff.FeatureFlag, error) {
	const updateFeatureFlagFmtStr = `
		UPDATE feature_flags
		SET
			flag_type = %s,
			bool_value = %s,
			rollout = %s
		WHERE flag_name = %s
		RETURNING
			flag_name,
			flag_type,
			bool_value,
			rollout,
			created_at,
			updated_at,
			deleted_at
		;
	`
	var (
		flagType string
		boolVal  *bool
		rollout  *int32
	)
	switch {
	case flag.Bool != nil:
		flagType = "bool"
		boolVal = &flag.Bool.Value
	case flag.Rollout != nil:
		flagType = "rollout"
		rollout = &flag.Rollout.Rollout
	default:
		return nil, errors.New("feature flag must have exactly one type")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		updateFeatureFlagFmtStr,
		flagType,
		boolVal,
		rollout,
		flag.Name,
	))
	return scanFeatureFlag(row)
}

func (f *featureFlagStore) DeleteFeatureFlag(ctx context.Context, name string) error {
	const deleteFeatureFlagFmtStr = `
		UPDATE feature_flags
		SET
			flag_name = flag_name || '-DELETED-' || TRUNC(random() * 1000000)::varchar(255),
			deleted_at = now()
		WHERE flag_name = %s;
	`

	err := f.Exec(ctx, sqlf.Sprintf(deleteFeatureFlagFmtStr, name))

	if err == nil {
		ff.ClearFlagFromCache(name)
	}

	return err
}

func (f *featureFlagStore) CreateRollout(ctx context.Context, name string, rollout int32) (*ff.FeatureFlag, error) {
	return f.CreateFeatureFlag(ctx, &ff.FeatureFlag{
		Name: name,
		Rollout: &ff.FeatureFlagRollout{
			Rollout: rollout,
		},
	})
}

func (f *featureFlagStore) CreateBool(ctx context.Context, name string, value bool) (*ff.FeatureFlag, error) {
	return f.CreateFeatureFlag(ctx, &ff.FeatureFlag{
		Name: name,
		Bool: &ff.FeatureFlagBool{
			Value: value,
		},
	})
}

var ErrInvalidColumnState = errors.New("encountered column that is unexpectedly null based on column type")

// rowScanner is an interface that can scan from either a sql.Row or sql.Rows
type rowScanner interface {
	Scan(...any) error
}

func scanFeatureFlag(scanner rowScanner) (*ff.FeatureFlag, error) {
	var (
		res      ff.FeatureFlag
		flagType string
		boolVal  *bool
		rollout  *int32
	)
	err := scanner.Scan(
		&res.Name,
		&flagType,
		&boolVal,
		&rollout,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	switch flagType {
	case "bool":
		if boolVal == nil {
			return nil, ErrInvalidColumnState
		}
		res.Bool = &ff.FeatureFlagBool{
			Value: *boolVal,
		}
	case "rollout":
		if rollout == nil {
			return nil, ErrInvalidColumnState
		}
		res.Rollout = &ff.FeatureFlagRollout{
			Rollout: *rollout,
		}
	default:
		return nil, ErrInvalidColumnState
	}

	return &res, nil
}

func (f *featureFlagStore) GetFeatureFlag(ctx context.Context, flagName string) (*ff.FeatureFlag, error) {
	const getFeatureFlagsQuery = `
		SELECT
			flag_name,
			flag_type,
			bool_value,
			rollout,
			created_at,
			updated_at,
			deleted_at
		FROM feature_flags
		WHERE deleted_at IS NULL
			AND flag_name = %s;
	`

	row := f.QueryRow(ctx, sqlf.Sprintf(getFeatureFlagsQuery, flagName))
	return scanFeatureFlag(row)
}

func (f *featureFlagStore) GetFeatureFlags(ctx context.Context) ([]*ff.FeatureFlag, error) {
	const listFeatureFlagsQuery = `
		SELECT
			flag_name,
			flag_type,
			bool_value,
			rollout,
			created_at,
			updated_at,
			deleted_at
		FROM feature_flags
		WHERE deleted_at IS NULL;
	`

	rows, err := f.Query(ctx, sqlf.Sprintf(listFeatureFlagsQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*ff.FeatureFlag, 0, 10)
	for rows.Next() {
		flag, err := scanFeatureFlag(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, flag)
	}
	return res, nil
}

func (f *featureFlagStore) CreateOverride(ctx context.Context, override *ff.Override) (*ff.Override, error) {
	const newFeatureFlagOverrideFmtStr = `
		INSERT INTO feature_flag_overrides (
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		) VALUES (
			%s,
			%s,
			%s,
			%s
		) RETURNING
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value;
	`
	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFeatureFlagOverrideFmtStr,
		&override.OrgID,
		&override.UserID,
		&override.FlagName,
		&override.Value))
	return scanFeatureFlagOverride(row)
}

func (f *featureFlagStore) DeleteOverride(ctx context.Context, orgID, userID *int32, flagName string) error {
	const newFeatureFlagOverrideFmtStr = `
		DELETE FROM feature_flag_overrides
		WHERE
			%s AND flag_name = %s;
	`

	var cond *sqlf.Query
	switch {
	case orgID != nil:
		cond = sqlf.Sprintf("namespace_org_id = %s", *orgID)
	case userID != nil:
		cond = sqlf.Sprintf("namespace_user_id = %s", *userID)
	default:
		return errors.New("must set either orgID or userID")
	}

	return f.Exec(ctx, sqlf.Sprintf(
		newFeatureFlagOverrideFmtStr,
		cond,
		flagName,
	))
}

func (f *featureFlagStore) UpdateOverride(ctx context.Context, orgID, userID *int32, flagName string, newValue bool) (*ff.Override, error) {
	const newFeatureFlagOverrideFmtStr = `
		UPDATE feature_flag_overrides
		SET flag_value = %s
		WHERE %s -- namespace condition
			AND flag_name = %s
		RETURNING
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value;
	`

	var cond *sqlf.Query
	switch {
	case orgID != nil:
		cond = sqlf.Sprintf("namespace_org_id = %s", *orgID)
	case userID != nil:
		cond = sqlf.Sprintf("namespace_user_id = %s", *userID)
	default:
		return nil, errors.New("must set either orgID or userID")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFeatureFlagOverrideFmtStr,
		newValue,
		cond,
		flagName,
	))
	return scanFeatureFlagOverride(row)
}

func (f *featureFlagStore) GetOverridesForFlag(ctx context.Context, flagName string) ([]*ff.Override, error) {
	const listFlagOverridesFmtString = `
		SELECT
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		FROM feature_flag_overrides
		WHERE flag_name = %s
			AND deleted_at IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listFlagOverridesFmtString, flagName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFeatureFlagOverrides(rows)
}

// GetUserOverrides lists the overrides that have been specifically set for the given userID.
// NOTE: this does not return any overrides for the user orgs. Those are returned separately
// by ListOrgOverridesForUser so they can be mered in proper priority order.
func (f *featureFlagStore) GetUserOverrides(ctx context.Context, userID int32) ([]*ff.Override, error) {
	const listUserOverridesFmtString = `
		SELECT
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		FROM feature_flag_overrides
		WHERE namespace_user_id = %s
			AND deleted_at IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFeatureFlagOverrides(rows)
}

// GetOrgOverridesForUser lists the feature flag overrides for all orgs the given user belongs to.
func (f *featureFlagStore) GetOrgOverridesForUser(ctx context.Context, userID int32) ([]*ff.Override, error) {
	const listUserOverridesFmtString = `
		SELECT
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		FROM feature_flag_overrides
		WHERE EXISTS (
			SELECT org_id
			FROM org_members
			WHERE org_members.user_id = %s
				AND feature_flag_overrides.namespace_org_id = org_members.org_id
		) AND deleted_at IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFeatureFlagOverrides(rows)
}

// GetOrgOverrideForFlag returns the flag override for the given organization.
func (f *featureFlagStore) GetOrgOverrideForFlag(ctx context.Context, orgID int32, flagName string) (*ff.Override, error) {
	const listOrgOverridesFmtString = `
		SELECT
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		FROM feature_flag_overrides
		WHERE namespace_org_id = %s
			AND flag_name = %s
			AND deleted_at IS NULL;
	`
	row := f.QueryRow(ctx, sqlf.Sprintf(listOrgOverridesFmtString, orgID, flagName))
	override, err := scanFeatureFlagOverride(row)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return override, nil
}

func scanFeatureFlagOverrides(rows *sql.Rows) ([]*ff.Override, error) {
	var res []*ff.Override
	for rows.Next() {
		override, err := scanFeatureFlagOverride(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, override)
	}
	return res, nil
}

func scanFeatureFlagOverride(scanner rowScanner) (*ff.Override, error) {
	var res ff.Override
	err := scanner.Scan(
		&res.OrgID,
		&res.UserID,
		&res.FlagName,
		&res.Value,
	)
	return &res, err
}

// GetUserFlags returns the calculated values for feature flags for the given userID. This should
// be the primary entrypoint for getting the user flags since it handles retrieving all the flags,
// the org overrides, and the user overrides, and merges them in priority order.
// TODO: remove
func (f *featureFlagStore) GetUserFlags(ctx context.Context, userID int32) (map[string]bool, error) {
	g, ctx := errgroup.WithContext(ctx)

	var flags []*ff.FeatureFlag
	g.Go(func() error {
		res, err := f.GetFeatureFlags(ctx)
		flags = res
		return err
	})

	var orgOverrides []*ff.Override
	g.Go(func() error {
		res, err := f.GetOrgOverridesForUser(ctx, userID)
		orgOverrides = res
		return err
	})

	var userOverrides []*ff.Override
	g.Go(func() error {
		res, err := f.GetUserOverrides(ctx, userID)
		userOverrides = res
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	res := make(map[string]bool, len(flags))
	for _, ff := range flags {
		res[ff.Name] = ff.EvaluateForUser(userID)

		// Org overrides are higher priority than default
		for _, oo := range orgOverrides {
			res[oo.FlagName] = oo.Value
		}

		// User overrides are higher priority than org overrides
		for _, uo := range userOverrides {
			res[uo.FlagName] = uo.Value
		}
	}

	return res, nil
}

// GetAnonymousUserFlags returns the calculated values for feature flags for the given anonymousUID
// TODO: remove
func (f *featureFlagStore) GetAnonymousUserFlags(ctx context.Context, anonymousUID string) (map[string]bool, error) {
	flags, err := f.GetFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]bool, len(flags))
	for _, ff := range flags {
		res[ff.Name] = ff.EvaluateForAnonymousUser(anonymousUID)
	}

	return res, nil
}

// TODO: remove
func (f *featureFlagStore) GetGlobalFeatureFlags(ctx context.Context) (map[string]bool, error) {
	flags, err := f.GetFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]bool, len(flags))
	for _, ff := range flags {
		if val, ok := ff.EvaluateGlobal(); ok {
			res[ff.Name] = val
		}
	}

	return res, nil
}

// GetUserFlags returns the calculated values for feature flags for the given userID. This should
// be the primary entrypoint for getting the user flags since it handles retrieving all the flags,
// the org overrides, and the user overrides, and merges them in priority order.
func (f *featureFlagStore) GetUserFlag(ctx context.Context, userID int32, flagName string) (*bool, error) {
	g, ctx := errgroup.WithContext(ctx)

	var flags []*ff.FeatureFlag
	g.Go(func() error {
		res, err := f.GetFeatureFlags(ctx)
		flags = res
		return err
	})

	var orgOverrides []*ff.Override
	g.Go(func() error {
		res, err := f.GetOrgOverridesForUser(ctx, userID)
		orgOverrides = res
		return err
	})

	var userOverrides []*ff.Override
	g.Go(func() error {
		res, err := f.GetUserOverrides(ctx, userID)
		userOverrides = res
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, ff := range flags {
		if ff.Name != flagName {
			res := ff.EvaluateForUser(userID)

			// Org overrides are higher priority than default
			for _, oo := range orgOverrides {
				res = oo.Value
			}

			// User overrides are higher priority than org overrides
			for _, uo := range userOverrides {
				res = uo.Value
			}
			return &res, nil
		}
	}

	return nil, errors.Errorf("Feature \"%s\" flag not found", flagName)
}

// GetAnonymousUserFlags returns the calculated values for feature flags for the given anonymousUID
func (f *featureFlagStore) GetAnonymousUserFlag(ctx context.Context, anonymousUID string, flagName string) (*bool, error) {
	flags, err := f.GetFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}

	for _, ff := range flags {
		if ff.Name != flagName {
			res := ff.EvaluateForAnonymousUser(anonymousUID)
			return &res, nil
		}
	}

	return nil, errors.Errorf("Feature \"%s\" flag not found", flagName)
}

func (f *featureFlagStore) GetGlobalFeatureFlag(ctx context.Context, flagName string) (*bool, error) {
	flags, err := f.GetFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}

	for _, ff := range flags {
		if ff.Name != flagName {
			if val, ok := ff.EvaluateGlobal(); ok {
				return &val, nil
			}
		}
	}

	return nil, errors.Errorf("Feature \"%s\" flag not found", flagName)
}

// GetOrgFeatureFlag returns the calculated flag value for the given organization, taking potential override into account
func (f *featureFlagStore) GetOrgFeatureFlag(ctx context.Context, orgID int32, flagName string) (bool, error) {
	g, ctx := errgroup.WithContext(ctx)

	var override *ff.Override
	var globalFlag *ff.FeatureFlag

	g.Go(func() error {
		res, err := f.GetOrgOverrideForFlag(ctx, orgID, flagName)
		override = res
		return err
	})
	g.Go(func() error {
		res, err := f.GetFeatureFlag(ctx, flagName)
		if err == sql.ErrNoRows {
			return nil
		}
		globalFlag = res
		return err
	})
	if err := g.Wait(); err != nil {
		return false, err
	}

	if override != nil {
		return override.Value, nil
	} else if globalFlag != nil {
		return globalFlag.Bool.Value, nil
	}

	return false, nil
}
