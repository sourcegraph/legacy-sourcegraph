package store

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
)

// CursorResultset builds the three value return tuple that is commonly returned
// from paginated query functions: the page of results, the cursor for the next
// page, and any error(s) that occurred.
//
// If an error is provided in err, it will be returned as is and no other
// processing will occur. This makes CursorResultset easier to use directly
// after a call to Store.query.
//
// If there are no further pages, the cursor will be set to 0. Users of this
// function must be careful to ensure that the field they are using for the
// cursor cannot be 0 in the normal course of events, otherwise weird things may
// happen! However, note that PostgreSQL SERIAL types start at 1, so this isn't
// a concern for normal ID fields.
func CursorResultset[T Cursor](o CursorOpts, results []T, err error) ([]T, int64, error) {
	if err != nil {
		return nil, 0, err
	}
	if o.Limit == 0 {
		return results, 0, nil
	}

	if len(results) > o.Limit {
		return results[0:o.Limit], results[o.Limit].Cursor(), nil
	} else if len(results) > 0 {
		return results, 0, nil
	}

	return nil, 0, nil
}

// CursorIntResultset is a specialised version of CursorResultset for cases
// where T is an integer type rather than a struct or interface, such as
// returning a resultset of IDs. Its behaviour is otherwise identical to
// CursorResultset.
func CursorIntResultset[T ~int | ~int8 | ~int16 | ~int32 | ~int64](o CursorOpts, results []T, err error) ([]T, int64, error) {
	if err != nil {
		return nil, 0, err
	}
	if o.Limit == 0 {
		return results, 0, nil
	}

	if len(results) > o.Limit {
		return results[0:o.Limit], int64(results[o.Limit]), nil
	} else if len(results) > 0 {
		return results, 0, nil
	}

	return nil, 0, nil
}

// CursorOpts allow code calling paginated list queries to specify how many
// items are desired in each page, along with any cursor that may be used.
//
// The first page should be retrieved by setting Cursor to 0, or simply leaving
// it unset.
type CursorOpts struct {
	Limit  int
	Cursor int64
}

// LimitDB provides the LIMIT clause for a paginated query.
func (o CursorOpts) LimitDB() *sqlf.Query {
	if o.Limit == 0 {
		return sqlf.Sprintf("")
	}
	return sqlf.Sprintf(fmt.Sprintf("LIMIT %d", o.Limit+1))
}

// WhereDB provides the WHERE clause for a paginated query. This must be AND-ed
// with any other clauses in the query.
//
// Note that the direction MUST match the ORDER BY clause in the query,
// otherwise extremely odd things will happen.
func (o CursorOpts) WhereDB(cursorField string, direction CursorDirection) *sqlf.Query {
	if o.Cursor == 0 {
		return sqlf.Sprintf("TRUE")
	}
	op := ">="
	if direction == CursorDirectionDescending {
		op = "<="
	}
	return sqlf.Sprintf("%s %s %s", sqlf.Sprintf(cursorField), sqlf.Sprintf(op), o.Cursor)
}

func (o CursorOpts) dbLimit() int {
	if o.Limit == 0 {
		return o.Limit
	}
	return o.Limit + 1
}

type Cursor interface {
	Cursor() int64
}

// CursorDirection indicates the direction of iteration through the paginated
// resultset. By default, it will be ascending.
//
// Query function implementors may choose to expose this in their public options
// API (in which case, see PaginationTest in cursor_test.go for an example of
// how to use the direction when building queries), or may have a hard coded
// direction, as appropriate.
type CursorDirection int

const (
	CursorDirectionAscending CursorDirection = iota
	CursorDirectionDescending
)

func (dir CursorDirection) String() string {
	if dir == CursorDirectionDescending {
		return "DESC"
	}
	return "ASC"
}
