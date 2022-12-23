package codeowners

import (
	"bufio"
	"io"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

// Parse parses CODEOWNERS file given as string and returns the proto
// representation of all rules within. The rules are in the same order
// as in the file, since this matters for evaluation.
func Parse(codeownersFile string) (*codeownerspb.File, error) {
	return Read(strings.NewReader(in))
}

// Read parses CODEOWNERS file given as a Reader and returns the proto
// representation of all rules within. The rules are in the same order
// as in the file, since this matters for evaluation.
func Read(in io.Reader) (*codeownerspb.File, error) {
	scanner := bufio.NewScanner(in)
	var rs []*codeownerspb.Rule
	p := new(parsing)
	for scanner.Scan() {
		p.nextLine(scanner.Text())
		if p.isBlank() {
			continue
		}
		if p.matchSection() {
			continue
		}
		pattern, owners, ok := p.matchRule()
		if !ok {
			return nil, errors.Errorf("failed to match rule: %s", p.line)
		}
		// Need to handle this error once, codeownerspb.File supports
		// error metadata.
		r := codeownerspb.Rule{
			Pattern:     unescape(pattern),
			SectionName: strings.TrimSpace(strings.ToLower(p.section)),
		}
		for _, ownerText := range owners {
			var o codeownerspb.Owner
			if strings.HasPrefix(ownerText, "@") {
				o.Handle = strings.TrimPrefix(ownerText, "@")
			} else {
				// Note: we assume owner text is an email if it does not
				// start with an `@` which would make it a handle.
				o.Email = ownerText
			}
			r.Owner = append(r.Owner, &o)
		}
		rs = append(rs, &r)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &codeownerspb.File{Rule: rs}, nil
}

// parsing implements matching and parsing primitives for CODEOWNERS files
// as well as keeps track of internal state as a file is being parsed.
type parsing struct {
	// line is the current line being parsed. CODEOWNERS files are built
	// in such a way that for syntactic purposes, every line can be considered
	// in isolation.
	line string
	// The most recently defined section, or "" if none.
	section string
}

// nextLine advances parsing to focus on the next line.
// Conveniently returns the same object for chaining with `notBlank()`.
func (p *parsing) nextLine(line string) {
	p.line = line
}

// rulePattern is expected to match a rule line like:
// `cmd/**/docs/index.md @readme-owners owner@example.com`.
//
//	^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//
// The first capturing   The second capturing group
// group extracts        extracts all the owners
// the file pattern.     separated by whitespace.
//
// The first capturing group supports escaping a whitespace with a `\`,
// so that the space is not treated as a separator between the pattern
// and owners.
var rulePattern = regexp.MustCompile(`^\s*((?:\\.|\S)+)((?:\s+\S+)*)\s*$`)

// matchRule tries to extract a codeowners rule from the current line
// and return the file pattern and one or more owners.
// Match is indicated by the third return value being true.
//
// Note: Need to check if a line matches a section using `matchSection`
// before matching a rule with this method, as `matchRule` will actually
// match a section line. This is because `matchRule` does not verify
// whether a pattern is a valid pattern. A line like "[documentation]"
// would be considered a pattern without owners (which is supported).
func (p *parsing) matchRule() (string, []string, bool) {
	match := rulePattern.FindStringSubmatch(p.lineWithoutComments())
	if len(match) != 3 {
		return "", nil, false
	}
	filePattern := match[1]
	owners := strings.Fields(match[2])
	return filePattern, owners, true
}

var sectionPattern = regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)

// matchSection tries to extract a section which looks like `[section name]`.
func (p *parsing) matchSection() bool {
	match := sectionPattern.FindStringSubmatch(p.lineWithoutComments())
	if len(match) != 2 {
		return false
	}
	p.section = match[1]
	return true
}

// isBlank returns true if the current line has no semantically relevant
// content. It can be blank while containing comments or whitespace.
func (p *parsing) isBlank() bool {
	return strings.TrimSpace(p.lineWithoutComments()) == ""
}

const (
	commentStart    = rune('#')
	escapeCharacter = rune('\\')
)

// lineWithoutComments returns the current line with the commented part
// stripped out.
func (p *parsing) lineWithoutComments() string {
	// A sensible default for index of the first byte where line-comment
	// starts is the line length. When the comment is removed by slicing
	// the string at the end, using the line-length as the index
	// of the first character dropped, yields the original string.
	commentStartIndex := len(p.line)
	var esc bool // whether current character is escaped.
	for i, c := range p.line {
		// Unescaped # seen - this is where the comment starts.
		if c == commentStart && !esc {
			commentStartIndex = i
			break
		}
		// Seeing escape character that is not being escaped itself (like \\)
		// means the following character is escaped.
		if c == escapeCharacter && !esc {
			esc = true
			continue
		}
		// Otherwise the next character is definitely not escaped.
		esc = false
	}
	return p.line[:commentStartIndex]
}

func unescape(s string) string {
	var b strings.Builder
	var esc bool // whether the current character is escaped
	for _, r := range s {
		if r == escapeCharacter && !esc {
			esc = true
			continue
		}
		b.WriteRune(r)
		esc = false
	}
	return b.String()
}
