package query

import (
	"errors"
	"fmt"
	"strings"
)

// LowercaseFieldNames performs strings.ToLower on every field name.
func LowercaseFieldNames(nodes []Node) []Node {
	return MapParameter(nodes, func(field, value string, negated bool) Node {
		return Parameter{Field: strings.ToLower(field), Value: value, Negated: negated}
	})
}

// HoistOr is a heuristic that rewrites simple but possibly ambiguous queries
// containing or-expressions. It changes certain expressions in a way that some
// consider to be more natural. For example, the following query without
// parentheses is interpreted as follows in the grammar:
//
// repo:foo a or b and c => (repo:foo a) or ((b) and (c))
//
// This function rewrites the above expression as follows:
//
// repo:foo a or b and c => repo:foo (a or b and c)
//
// Any number of field:value parameters may occur before and after the pattern
// expression separated by or-operators, and these are hoisted out. The pattern
// expression must be contiguous. If not, we want to preserve the default
// interpretation, which corresponds more naturally to groupings with field
// parameters, i.e.,
//
// repo:foo a or b or repo:bar c => (repo:foo a) or (b) or (repo:bar c)
func HoistOr(nodes []Node) ([]Node, error) {
	if len(nodes) != 1 {
		return nil, fmt.Errorf("heuristic requires one top-level")
	}

	expression, ok := nodes[0].(Operator)
	if !ok || expression.Kind != Or {
		return nil, fmt.Errorf("heuristic requires top-level or-expression")
	}

	n := len(expression.Operands)
	var pattern []Node
	var scopeParameters []Node
	for i, node := range expression.Operands {
		if i == 0 || i == n-1 {
			scopePart, patternPart, err := PartitionSearchPattern([]Node{node})
			if err != nil || patternPart == nil {
				return nil, errors.New("could not partition first or last expression")
			}
			pattern = append(pattern, patternPart)
			scopeParameters = append(scopeParameters, scopePart...)
			continue
		}
		if !isPatternExpression([]Node{node}) {
			return nil, fmt.Errorf("inner expression %s is not a pure pattern expression", node.String())
		}
		pattern = append(pattern, node)
	}
	return append(scopeParameters, newOperator(pattern, Or)...), nil
}
