package shared

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Location is an LSP-like location scoped to a dump.
type Location struct {
	UploadID int
	Path     core.UploadRelPath
	Range    Range
}

// Usage represents a def/ref/impl/super of some range/symbol that
// was queried for using 'usagesForSymbol' or similar.
type Usage struct {
	UploadID int
	// Path is the path of this usage wrt the
	// root of the scip.Index this usage was found in.
	Path  core.UploadRelPath
	Range Range
	// Symbol is the SCIP symbol at the _usage site_, which may not
	// be equal to the symbol at the _lookup site_. For the various
	// cases, see codeintel.codenav.graphql as well as extractOccurrenceData.
	Symbol string
	Kind   UsageKind
}

func (u Usage) ToLocation() Location {
	return Location{UploadID: u.UploadID, Path: u.Path, Range: u.Range}
}

// UsageKind is a more compact representation for SymbolUsageKind
// in the GraphQL API
type UsageKind int32

const (
	UsageKindDefinition UsageKind = iota
	UsageKindReference
	UsageKindImplementation
	UsageKindSuper
)

// RangesColumnName represents the column name in the codeintel_scip_symbols table
// that should be used when searching for a certain kind of usage.
func (k UsageKind) RangesColumnName() string {
	switch k {
	case UsageKindDefinition:
		return "definition_ranges"
	case UsageKindReference:
		return "reference_ranges"
	case UsageKindImplementation:
		return "implementation_ranges"
	case UsageKindSuper:
		return "definition_ranges"
		// For supers, we're looking for definitions of interfaces/super-class methods.
	default:
		panic(fmt.Sprintf("unhandled case for UsageKind: %v", k))
	}
}

func (k UsageKind) String() string {
	switch k {
	case UsageKindDefinition:
		return "definition"
	case UsageKindReference:
		return "reference"
	case UsageKindImplementation:
		return "implementation"
	case UsageKindSuper:
		return "super"
	default:
		panic(fmt.Sprintf("unhandled case for UsageKind: %d", int32(k)))
	}
}

// UsageBuilder is a transient type representing some Usage
// that will be constructed in the future, but it's not yet clear what
// the Kind value ought to be for the Usage.
type UsageBuilder struct {
	Range  scip.Range
	Symbol string
	// SymbolRoles represents the 'role' of the underlying occurrence
	// which was used to construct this UsageBuilder.
	//
	// TODO: Make this field private after removing extractLocationsFromPosition
	SymbolRoles scip.SymbolRole
}

func (ub UsageBuilder) RangeKey() [4]int32 {
	return [4]int32{ub.Range.Start.Line, ub.Range.Start.Character, ub.Range.End.Line, ub.Range.End.Character}
}

func (ub UsageBuilder) SymbolRoleKey() int32 {
	return int32(ub.SymbolRoles)
}

func (ub UsageBuilder) SymbolAndRoleKey() string {
	return fmt.Sprintf("%s:%x", ub.Symbol, ub.SymbolRoles)
}

// Diagnostic describes diagnostic information attached to a location within a
// particular dump.
type Diagnostic[PathType any] struct {
	UploadID int
	Path     PathType
	precise.DiagnosticData
}

func AdjustDiagnostic(d Diagnostic[core.UploadRelPath], upload shared.CompletedUpload) Diagnostic[core.RepoRelPath] {
	return Diagnostic[core.RepoRelPath]{
		UploadID:       d.UploadID,
		Path:           core.NewRepoRelPath(upload, d.Path),
		DiagnosticData: d.DiagnosticData,
	}
}

// CodeIntelligenceRange pairs a range with its definitions, references, implementations, and hover text.
type CodeIntelligenceRange struct {
	Range           Range
	Definitions     []Location
	References      []Location
	Implementations []Location
	HoverText       string
}

// UploadLocation is a path and range pair from within a particular upload. The target commit
// denotes the target commit for which the location was set (the originally requested commit).
type UploadLocation struct {
	Upload       shared.CompletedUpload
	Path         core.RepoRelPath
	TargetCommit string
	TargetRange  Range
}

type SnapshotData struct {
	DocumentOffset int
	Symbol         string
	AdditionalData []string
}

type Range struct {
	Start Position
	End   Position
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
}

func (p Position) ToSCIPPosition() scip.Position {
	return scip.Position{
		Line:      int32(p.Line),
		Character: int32(p.Character),
	}
}

func TranslatePosition(r scip.Position) Position {
	return Position{
		Line:      int(r.Line),
		Character: int(r.Character),
	}
}

func NewRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{
		Start: Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

func TranslateRange(r scip.Range) Range {
	return NewRange(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character))
}

func (r Range) ToSCIPRange() scip.Range {
	return scip.NewRangeUnchecked([]int32{
		int32(r.Start.Line), int32(r.Start.Character),
		int32(r.End.Line), int32(r.End.Character),
	})
}

type Matcher struct {
	exactSymbol string
	start       scip.Position
	end         core.Option[scip.Position]
}

func NewStartPositionMatcher(start scip.Position) Matcher {
	return Matcher{start: start, exactSymbol: "", end: core.None[scip.Position]()}
}

// NewSCIPBasedMatcher creates a matcher based on the given range_.
//
// range_ should correspond to a single occurrence, not any arbitrary range.
// range_ must be well-formed.
func NewSCIPBasedMatcher(range_ scip.Range, exactSymbol string) Matcher {
	return Matcher{
		exactSymbol: exactSymbol,
		start:       range_.Start,
		end:         core.Some(range_.End),
	}
}

func (m *Matcher) Attrs() []attribute.KeyValue {
	var rangeStr string
	if end, ok := m.end.Get(); ok {
		rangeStr = fmt.Sprintf("[%d:%d, %d:%d)", m.start.Line, m.start.Character, end.Line, end.Character)
	} else {
		rangeStr = fmt.Sprintf("pos %d:%d", m.start.Line, m.start.Character)
	}
	return []attribute.KeyValue{
		attribute.String("matcher.symbol", m.exactSymbol),
		attribute.String("matcher.range", rangeStr),
	}
}

func (m *Matcher) PositionBased() (scip.Position, bool) {
	if m.end.IsSome() {
		return scip.Position{}, false
	}
	return m.start, true
}

func (m *Matcher) SymbolBased() (string, scip.Range, bool) {
	if end, ok := m.end.Get(); ok {
		return m.exactSymbol, scip.Range{Start: m.start, End: end}, true
	}
	return "", scip.Range{}, false
}
