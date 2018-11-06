// Package query provides an expression tree structure which can be converted
// into WHERE queries. It is used by DB APIs to expose a more powerful query
// interface.
package query

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
)

// Q is a query item. It is converted into a *sqlf.Query by Eval.
type Q interface{}

// And returns a Q which when evaluated will join the children by "AND".
func And(children ...Q) Q {
	return &and{Children: children}
}

// Or returns a Q which when evaluated will join the children by "OR".
func Or(children ...Q) Q {
	return &or{Children: children}
}

// Not returns a Q which when evaluated will wrap child with "NOT".
func Not(child Q) Q {
	return &not{Child: child}
}

type and struct {
	Children []Q
}

type or struct {
	Children []Q
}

type not struct {
	Child Q
}

// Eval runs all atoms of q through queryFn, returning the final query to
// run. If any call of queryFn returns an error, that error is returned by
// Eval.
//
// Eval handles And, Or, Not and booleans. Otherwise every other Q will be
// passed to queryFn.
func Eval(q Q, queryFn func(q Q) (*sqlf.Query, error)) (*sqlf.Query, error) {
	childQueries := func(qs []Q) ([]*sqlf.Query, error) {
		x := make([]*sqlf.Query, 0, len(qs))
		for _, q := range qs {
			c, err := Eval(q, queryFn)
			if err != nil {
				return nil, err
			}
			x = append(x, c)
		}
		return x, nil
	}

	switch c := q.(type) {
	case *and:
		children, err := childQueries(c.Children)
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			return sqlf.Sprintf("TRUE"), nil
		}
		return sqlf.Sprintf("(%s)", sqlf.Join(children, "AND")), nil

	case *or:
		children, err := childQueries(c.Children)
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			return sqlf.Sprintf("FALSE"), nil
		}
		return sqlf.Sprintf("(%s)", sqlf.Join(children, "OR")), nil

	case *not:
		child, err := Eval(c.Child, queryFn)
		if err != nil {
			return nil, err
		}
		return sqlf.Sprintf("NOT(%s)", child), nil

	case bool:
		if c {
			return sqlf.Sprintf("TRUE"), nil
		}
		return sqlf.Sprintf("FALSE"), nil

	default:
		return queryFn(q)
	}
}

// Print returns a string representing Q.
func Print(q Q) string {
	rq, _ := Eval(q, func(q Q) (*sqlf.Query, error) {
		return sqlf.Sprintf("%s", q), nil
	})
	return fmt.Sprintf(rq.Query(printfBindVar{}), rq.Args()...)
}

type printfBindVar struct{}

func (printfBindVar) BindVar(i int) string {
	return "%q"
}
