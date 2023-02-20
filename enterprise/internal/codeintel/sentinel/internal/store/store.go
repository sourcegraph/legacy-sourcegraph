package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	Foo(ctx context.Context) error
	InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) error
	ScanMatches(ctx context.Context) error
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new sentinel store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("sentinel.store", ""),
		operations: newOperations(observationCtx),
	}
}

func (s *store) Foo(ctx context.Context) (err error) {
	// TODO
	return nil
}

func (s *store) InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) (err error) {
	ctx, _, endObservation := s.operations.insertVulnerabilities.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesTemporaryVulnerabilitiesTableQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesTemporaryVulnerabilityAffectedPackagesTableQuery)); err != nil {
		return err
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_vulnerabilities",
		batch.MaxNumPostgresParameters,
		[]string{
			"source_id",
			"summary",
			"details",
			"cpes",
			"cwes",
			"aliases",
			"related",
			"data_source",
			"urls",
			"severity",
			"cvss_vector",
			"cvss_score",
			"published",
			"modified",
			"withdrawn",
		},
		func(inserter *batch.Inserter) error {
			for _, v := range vulnerabilities {
				if v.CPEs == nil {
					v.CPEs = []string{}
				}
				if v.CWEs == nil {
					v.CWEs = []string{}
				}
				if v.Aliases == nil {
					v.Aliases = []string{}
				}
				if v.Related == nil {
					v.Related = []string{}
				}
				if v.URLs == nil {
					v.URLs = []string{}
				}

				if err := inserter.Insert(
					ctx,
					v.SourceID,
					v.Summary,
					v.Details,
					v.CPEs,
					v.CWEs,
					v.Aliases,
					v.Related,
					v.DataSource,
					v.URLs,
					v.Severity,
					v.CVSSVector,
					v.CVSSScore,
					v.Published,
					v.Modified,
					v.Withdrawn,
				); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
		return err
	}

	if err := batch.WithInserter(
		ctx,
		tx.Handle(),
		"t_vulnerability_affected_packages",
		batch.MaxNumPostgresParameters,
		[]string{
			"source_id",
			"package_name",
			"language",
			"namespace",
			"version_constraint",
			"fixed",
			"fixed_in",
			"affected_symbols",
		},
		func(inserter *batch.Inserter) error {
			for _, v := range vulnerabilities {
				for _, ap := range v.AffectedPackages {
					if ap.VersionConstraint == nil {
						ap.VersionConstraint = []string{}
					}
					if ap.AffectedSymbols == nil {
						ap.AffectedSymbols = []shared.AffectedSymbol{}
					}

					serialized, err := json.Marshal(ap.AffectedSymbols)
					if err != nil {
						return err
					}

					if err := inserter.Insert(
						ctx,
						v.SourceID,
						ap.PackageName,
						ap.Language,
						ap.Namespace,
						ap.VersionConstraint,
						ap.Fixed,
						ap.FixedIn,
						serialized,
					); err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesUpdateQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesAffectedPackagesUpdateQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(insertVulnerabilitiesAffectedSymbolsUpdateQuery)); err != nil {
		return err
	}

	return nil
}

const insertVulnerabilitiesTemporaryVulnerabilitiesTableQuery = `
CREATE TEMPORARY TABLE t_vulnerabilities (
	source_id    TEXT NOT NULL,
	summary      TEXT NOT NULL,
	details      TEXT NOT NULL,
	cpes         TEXT[] NOT NULL,
	cwes         TEXT[] NOT NULL,
	aliases      TEXT[] NOT NULL,
	related      TEXT[] NOT NULL,
	data_source  TEXT NOT NULL,
	urls         TEXT[] NOT NULL,
	severity     TEXT NOT NULL,
	cvss_vector  TEXT NOT NULL,
	cvss_score   TEXT NOT NULL,
	published    TIMESTAMP WITH TIME ZONE,
	modified     TIMESTAMP WITH TIME ZONE NOT NULL,
	withdrawn    TIMESTAMP WITH TIME ZONE NOT NULL
) ON COMMIT DROP
`

const insertVulnerabilitiesTemporaryVulnerabilityAffectedPackagesTableQuery = `
CREATE TEMPORARY TABLE t_vulnerability_affected_packages (
	source_id           TEXT NOT NULL,
	package_name        TEXT NOT NULL,
	language            TEXT NOT NULL,
	namespace           TEXT NOT NULL,
	version_constraint  TEXT[] NOT NULL,
	fixed               boolean NOT NULL,
	fixed_in            TEXT NOT NULL,
	affected_symbols    JSON NOT NULL
) ON COMMIT DROP
`

const insertVulnerabilitiesUpdateQuery = `
INSERT INTO vulnerabilities (
	source_id,
	summary,
	details,
	cpes,
	cwes,
	aliases,
	related,
	data_source,
	urls,
	severity,
	cvss_vector,
	cvss_score,
	published,
	modified,
	withdrawn
)
SELECT
	source_id,
	summary,
	details,
	cpes,
	cwes,
	aliases,
	related,
	data_source,
	urls,
	severity,
	cvss_vector,
	cvss_score,
	published,
	modified,
	withdrawn
FROM t_vulnerabilities

-- TODO - update instead
ON CONFLICT DO NOTHING
`

const insertVulnerabilitiesAffectedPackagesUpdateQuery = `
INSERT INTO vulnerability_affected_packages(
	vulnerability_id,
	package_name,
	language,
	namespace,
	version_constraint,
	fixed,
	fixed_in
)
SELECT
	(SELECT v.id FROM vulnerabilities v WHERE v.source_id = vap.source_id),
	package_name,
	language,
	namespace,
	version_constraint,
	fixed,
	fixed_in
FROM t_vulnerability_affected_packages vap

-- TODO - update instead
ON CONFLICT DO NOTHING
`

const insertVulnerabilitiesAffectedSymbolsUpdateQuery = `
WITH
json_candidates AS (
	SELECT
		vap.id,
		json_array_elements(tvap.affected_symbols) AS affected_symbol
	FROM t_vulnerability_affected_packages tvap
	JOIN vulnerability_affected_packages vap ON vap.package_name = tvap.package_name
	JOIN vulnerabilities v ON v.id = vap.vulnerability_id
	WHERE
		v.source_id = tvap.source_id
),
candidates AS (
	SELECT
		c.id,
		c.affected_symbol->'path'::text AS path,
		ARRAY(SELECT json_array_elements_text(c.affected_symbol->'symbols'))::text[] AS symbols
	FROM json_candidates c
)
INSERT INTO vulnerability_affected_symbols(vulnerability_affected_package_id, path, symbols)
SELECT c.id, c.path, c.symbols FROM candidates c

-- TODO - update instead
ON CONFLICT DO NOTHING
`

func (s *store) ScanMatches(ctx context.Context) (err error) {
	ctx, _, endObservation := s.operations.scanMatches.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO - insert instead of returning
	matches, err := scanFilteredVulnerabilityMatches(s.db.Query(ctx, sqlf.Sprintf(scanMatchesQuery)))
	if err != nil {
		return err
	}

	fmt.Printf("> %v\n", matches)
	return nil
}

const scanMatchesQuery = `
SELECT
	r.dump_id,
	r.version,
	vap.vulnerability_id,
	vap.version_constraint
FROM vulnerability_affected_packages vap
JOIN lsif_references r ON r.name LIKE vap.package_name
WHERE
	-- TODO - java mapping
	-- r.scheme = 'gomod' AND vap.language = 'go'
	r.scheme = 'npm' AND vap.language = 'Javascript'
`

type VulnerabilityMatch struct {
	UploadID        int
	VulnerabilityID int
}

var scanFilteredVulnerabilityMatches = basestore.NewFilteredSliceScanner(func(s dbutil.Scanner) (m VulnerabilityMatch, _ bool, _ error) {
	var version string
	var versionConstraints []string

	if err := s.Scan(&m.UploadID, &version, &m.VulnerabilityID, pq.Array(&versionConstraints)); err != nil {
		return VulnerabilityMatch{}, false, err
	}

	return m, versionMatchesConstraints(version, versionConstraints), nil
})

func versionMatchesConstraints(version string, constraints []string) bool {
	// TODO - actually compare constraints
	fmt.Printf("> %v %v\n", version, constraints)
	return true
}
