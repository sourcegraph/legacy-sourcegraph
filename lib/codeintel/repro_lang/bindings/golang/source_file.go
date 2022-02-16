package repro_lang

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type reproSourceFile struct {
	Source      *lsif_typed.SourceFile
	node        *sitter.Node
	definitions []*definitionStatement
	references  []*referenceStatement
	localScope  *reproScope
}

func newSourceFile(sourceFile *lsif_typed.SourceFile, node *sitter.Node) *reproSourceFile {
	return &reproSourceFile{
		Source:      sourceFile,
		node:        node,
		definitions: nil,
		references:  nil,
		localScope:  newScope(),
	}
}

func (d *reproSourceFile) slicePosition(n *sitter.Node) string {
	return d.Source.Text[n.StartByte():n.EndByte()]
}
func (d *reproSourceFile) newIdentifier(n *sitter.Node) *identifier {
	if n == nil {
		return nil
	}
	if n.Type() != "identifier" {
		panic("expected identifier, obtained " + n.Type())
	}
	value := d.slicePosition(n)
	globalIdentifier := n.ChildByFieldName("global")
	if globalIdentifier != nil {
		projectName := globalIdentifier.ChildByFieldName("project_name")
		descriptors := globalIdentifier.ChildByFieldName("descriptors")
		if projectName != nil && descriptors != nil {
			value = fmt.Sprintf("global %v %v", d.slicePosition(projectName), d.slicePosition(descriptors))
		}
	}
	return &identifier{
		value:    value,
		position: NewRangePositionFromTreeSitter(n),
	}
}

func NewRangePositionFromTreeSitter(node *sitter.Node) *lsif_typed.Range {
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
