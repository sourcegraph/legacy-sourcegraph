package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// UpdatePackageReferences inserts reference data tied to the given upload.
func (s *Store) UpdatePackageReferences(ctx context.Context, dumpID int, references []semantic.PackageReference) (err error) {
	ctx, endObservation := s.operations.updatePackageReferences.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numReferences", len(references)),
	}})
	defer endObservation(1, observation.Args{})

	if len(references) == 0 {
		return nil
	}

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_references without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(updateReferencesTemporaryTableQuery)); err != nil {
		return err
	}

	db := tx.Handle().DB()
	columns := []string{"scheme", "name", "version", "filter"}

	// Bulk insert all the unique column values into the temporary table
	if err := batch.InsertValues(ctx, db, "t_lsif_references", columns, loadReferencesChannel(references)); err != nil {
		return err
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized idump id here since it is the same for all rows in this operation.
	return tx.Exec(ctx, sqlf.Sprintf(updateReferencesInsertQuery, dumpID))
}

const updateReferencesTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/references.go:UpdatePackageReferences
CREATE TEMPORARY TABLE t_lsif_references (
	scheme text NOT NULL,
	name text NOT NULL,
	version text NOT NULL,
	filter bytea NOT NULL
) ON COMMIT DROP
`

const updateReferencesInsertQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/references.go:UpdatePackageReferences
INSERT INTO lsif_references (dump_id, scheme, name, version, filter)
SELECT %s, source.scheme, source.name, source.version, source.filter
FROM t_lsif_references source
`

func loadReferencesChannel(references []semantic.PackageReference) <-chan []interface{} {
	ch := make(chan []interface{}, len(references))

	go func() {
		defer close(ch)

		for _, r := range references {
			filter := r.Filter
			if filter == nil {
				// avoid not null constraint
				filter = []byte{}
			}

			ch <- []interface{}{r.Scheme, r.Name, r.Version, filter}
		}
	}()

	return ch
}
