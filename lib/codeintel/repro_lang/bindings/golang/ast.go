package repro_lang

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type definitionStatement struct {
	docstring           string
	name                *identifier
	implementsRelation  *identifier
	referencesRelation  *identifier
	typeDefinesRelation *identifier
}

func (s *definitionStatement) relationIdentifiers() []*identifier {
	return []*identifier{s.implementsRelation, s.referencesRelation, s.typeDefinesRelation}
}

type referenceStatement struct {
	name *identifier
}

type identifier struct {
	value    string
	symbol   string
	position *lsif_typed.Range
}

func newIdentifier(s *reproSourceFile, n *sitter.Node) *identifier {
	if n == nil {
		return nil
	}
	if n.Type() != "identifier" {
		panic("expected identifier, obtained " + n.Type())
	}
	value := s.nodeText(n)
	globalIdentifier := n.ChildByFieldName("global")
	if globalIdentifier != nil {
		projectName := globalIdentifier.ChildByFieldName("project_name")
		descriptors := globalIdentifier.ChildByFieldName("descriptors")
		if projectName != nil && descriptors != nil {
			value = fmt.Sprintf("global %v %v", s.nodeText(projectName), s.nodeText(descriptors))
		}
	}
	return &identifier{
		value:    value,
		position: NewRangePositionFromNode(n),
	}
}

func NewRangePositionFromNode(node *sitter.Node) *lsif_typed.Range {
	return &lsif_typed.Range{
		Start: lsif_typed.Position{
			Line:      int(node.StartPoint().Row),
			Character: int(node.StartPoint().Column),
		},
		End: lsif_typed.Position{
			Line:      int(node.EndPoint().Row),
			Character: int(node.EndPoint().Column),
		},
	}
}

func (i *identifier) resolveSymbol(localScope *reproScope, context *reproContext) {
	scope := context.globalScope
	if i.isLocalSymbol() {
		scope = localScope
	}
	symbol, ok := scope.names[i.value]
	if !ok {
		if i.value == "global global-workspace hello.repro/hello()." {
			fmt.Println("scope", context.globalScope)
		}
		symbol = "local ERROR_UNRESOLVED_SYMBOL"
	}
	i.symbol = symbol
}

func (i *identifier) isLocalSymbol() bool {
	return strings.HasPrefix(i.value, "local")
}
