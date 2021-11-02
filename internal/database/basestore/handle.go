package basestore

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// TransactableHandle is a wrapper around a database connection that provides nested transactions
// through registration and finalization of savepoints. A transactable database handler can be
// shared by multiple stores.
type TransactableHandle struct {
	db         dbutil.DB
	savepoints []*savepoint
	txOptions  sql.TxOptions
}

// NewHandleWithDB returns a new transactable database handle using the given database connection.
func NewHandleWithDB(db dbutil.DB, txOptions sql.TxOptions) *TransactableHandle {
	return &TransactableHandle{db: db, txOptions: txOptions}
}

// DB returns the underlying database handle.
func (h *TransactableHandle) DB() dbutil.DB {
	return h.db
}

// InTransaction returns true if the underlying database handle is in a transaction.
func (h *TransactableHandle) InTransaction() bool {
	db := h.db
	if unwrapper, ok := db.(dbutil.Unwrapper); ok {
		// Unwrap in case the dbutil.DB is a database.db instead of *sql.DB or
		// *sql.Tx.  This is needed because below, we attempt to cast the db as
		// a dbutil.Tx, which is not implemented by database.db. This should
		// eventually be removed once dbutil.DB is subsumed by database.DB
		db = unwrapper.Unwrap()
	}

	_, ok := db.(dbutil.Tx)
	return ok
}

// Transact returns a new transactional database handle whose methods operate within the context of
// a new transaction or a new savepoint. This method will return an error if the underlying connection
// cannot be interface upgraded to a TxBeginner.
//
// Note that it is not valid to call Transact or Done on the same database handle from distinct goroutines.
// Because we support properly nested transactions via savepoints, calling Transact from two different
// goroutines on the same handle will not be deterministic: either transaction could nest the other one,
// and calling Done in one goroutine may not finalize the expected unit of work.
func (h *TransactableHandle) Transact(ctx context.Context) (*TransactableHandle, error) {
	db := h.db
	if unwrapper, ok := db.(dbutil.Unwrapper); ok {
		// Unwrap in case the dbutil.DB is a database.db instead of *sql.DB or
		// *sql.Tx.  This is needed because below, we attempt to cast the db as
		// a dbutil.TxBeginner, which is not implemented by database.db. This
		// should eventually be removed once dbutil.DB is subsumed by
		// database.DB
		db = unwrapper.Unwrap()
	}

	if h.InTransaction() {
		savepoint, err := newSavepoint(ctx, db)
		if err != nil {
			return nil, err
		}

		h.savepoints = append(h.savepoints, savepoint)
		return h, nil
	}

	tb, ok := db.(dbutil.TxBeginner)
	if !ok {
		return nil, ErrNotTransactable
	}

	tx, err := tb.BeginTx(ctx, &h.txOptions)
	if err != nil {
		return nil, err
	}

	return &TransactableHandle{db: tx, txOptions: h.txOptions}, nil
}

// Done performs a commit or rollback of the underlying transaction/savepoint depending
// on the value of the error parameter. The resulting error value is a multierror containing
// the error parameter along with any error that occurs during commit or rollback of the
// transaction/savepoint. If the store does not wrap a transaction the original error value
// is returned unchanged.
func (h *TransactableHandle) Done(err error) error {
	db := h.db
	if unwrapper, ok := db.(dbutil.Unwrapper); ok {
		// Unwrap in case the dbutil.DB is a database.db instead of *sql.DB or *sql.Tx
		// This is needed because below, we try to cast db as dbutil.Tx, which
		// is not implemented by database.db. This should eventually be removed once
		// dbutil.DB is subsumed by database.DB
		db = unwrapper.Unwrap()
	}

	if n := len(h.savepoints); n > 0 {
		var savepoint *savepoint
		savepoint, h.savepoints = h.savepoints[n-1], h.savepoints[:n-1]

		if err == nil {
			return savepoint.Commit()
		}
		return combineErrors(err, savepoint.Rollback())
	}

	tx, ok := db.(dbutil.Tx)
	if !ok {
		return err
	}

	if err == nil {
		return tx.Commit()
	}
	return combineErrors(err, tx.Rollback())
}

// combineErrors returns a multierror containing all fo the non-nil error parameter values.
// This method should be used over multierror when it is not guaranteed that the original
// error was non-nil (multierror.Append creates a non-nil error even if it is empty).
func combineErrors(errs ...error) (err error) {
	for _, e := range errs {
		if e != nil {
			if err == nil {
				err = e
			} else {
				err = multierror.Append(err, e)
			}
		}
	}

	return err
}
