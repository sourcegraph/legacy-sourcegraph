package d2ir

import (
	"fmt"

	"oss.terrastruct.com/d2/d2parser"
)

// QueryAll is only for tests and debugging.
func (m *Map) QueryAll(idStr string) (na []Node, _ error) {
	k, err := d2parser.ParseMapKey(idStr)
	if err != nil {
		return nil, err
	}

	if k.Key != nil {
		fa, err := m.EnsureField(k.Key, nil, false, nil)
		if err != nil {
			return nil, err
		}
		if len(fa) == 0 {
			return nil, nil
		}
		for _, f := range fa {
			if len(k.Edges) == 0 {
				na = append(na, f)
				return na, nil
			}
			m = f.Map()
			if m == nil {
				return nil, nil
			}
		}
	}

	eida := NewEdgeIDs(k)

	for i, eid := range eida {
		refctx := &RefContext{
			Key:      k,
			ScopeMap: m,
			Edge:     k.Edges[i],
		}
		ea := m.GetEdges(eid, refctx, nil)
		for _, e := range ea {
			if k.EdgeKey == nil {
				na = append(na, e)
			} else if e.Map_ != nil {
				f := e.Map_.GetField(k.EdgeKey.IDA()...)
				if f != nil {
					na = append(na, f)
				}
			}
		}
	}
	return na, nil
}

// Query is only for tests and debugging.
func (m *Map) Query(idStr string) (Node, error) {
	na, err := m.QueryAll(idStr)
	if err != nil {
		return nil, err
	}

	if len(na) == 0 {
		return nil, nil
	}
	if len(na) > 1 {
		return nil, fmt.Errorf("expected only one query result but got: %#v", na)
	}
	return na[0], nil
}
