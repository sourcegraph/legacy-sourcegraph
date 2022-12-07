package database

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoVersionsStore interface {
	CreateIfNotExists(ctx context.Context, version types.RepoVersion) (*types.RepoVersion, error)
	Lookup(ctx context.Context, repoID api.RepoID, externalID string) (*types.RepoVersion, error)
}

var _ RepoVersionsStore = (*repoVersionsStore)(nil)

// repoVersionsStore handles access to the repo_versions table
type repoVersionsStore struct {
	logger log.Logger
	*basestore.Store
}

func RepoVersionsWith(logger log.Logger, other basestore.ShareableStore) RepoVersionsStore {
	return &repoVersionsStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

// CreateIfNotExists ensures given version exists in the database after execution.
// Version must be uniquely identified by <repo ID, external ID>.
// Version passed in must not have an ID set.
func (s *repoVersionsStore) CreateIfNotExists(ctx context.Context, v types.RepoVersion) (*types.RepoVersion, error) {
	var id int
	encodedReachability, err := json.Marshal(v.Reachability)
	if err != nil {
		return nil, err
	}
	if err := s.Handle().QueryRowContext(
		ctx,
		`INSERT INTO repo_versions(repo_id, external_id, path_cover_color, path_cover_index, path_cover_reachability, updated_at, created_at)
		VALUES($1, $2, $3, $4, $5, now(), now())
		ON CONFLICT ("repo_id", "external_id") DO NOTHING
		RETURNING id`,
		v.RepoID, v.ExternalID, v.PathCoverage.PathColor, v.PathCoverage.PathIndex, string(encodedReachability),
	).Scan(&id); err != nil {
		return nil, err
	}
	// TODO this assumes there was no conflict which is incorrect:
	v.ID = id
	return &v, nil
}

func (s *repoVersionsStore) Lookup(ctx context.Context, repoID api.RepoID, externalID string) (*types.RepoVersion, error) {
	var v types.RepoVersion
	var reachability string
	err := s.Handle().QueryRowContext(
		ctx,
		`SELECT v.id, v.repo_id, v.external_id, v.path_cover_color, v.path_cover_index, v.path_cover_reachability
		FROM repo_versions AS v
		WHERE v.repo_id = $1
		AND v.external_id = $2`,
		repoID, externalID,
	).Scan(&v.ID, &v.RepoID, &v.ExternalID, &v.PathCoverage.PathColor, &v.PathCoverage.PathIndex, &reachability)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(reachability), &v.Reachability); err != nil {
		return nil, err
	}
	return &v, nil
}
