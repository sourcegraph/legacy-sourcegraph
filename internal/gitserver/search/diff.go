package search

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

const fileSeparator = " "

// FormattedDiff is a formatted diff between a commit and its parent in the format
// as generated by FormatDiff(). It is iterable in a zero-copy manner for searching
// its content.
type FormattedDiff string

// FormatDiff generates a formatted diff from git's patch output
// in the following structure:
//
// oldFile  newFile
// @@ hunk header
//  context line
// -removed line
// +added line
func FormatDiff(rawDiff []byte) FormattedDiff {
	var buf strings.Builder
	buf.Grow(len(rawDiff) / 2) // Rough (mostly lower bound) estimate

	lines := bytes.SplitAfter(rawDiff, []byte{'\n'})
	const (
		STATE_DELTA = iota
		STATE_HUNK
		STATE_LINE
	)

	state := STATE_DELTA

	var oldFile, newFile []byte
	for i := 0; i < len(lines); {
		line := lines[i]
		switch state {
		case STATE_DELTA:
			if bytes.HasPrefix(line, []byte("---")) {
				oldFile = line[len("--- ") : len(line)-1]
				i++
			} else if bytes.HasPrefix(line, []byte("+++")) {
				newFile = line[len("+++ ") : len(line)-1]
				i++
			} else if bytes.HasPrefix(line, []byte("@@")) {
				buf.Write(oldFile)
				buf.WriteString("  ")
				buf.Write(newFile)
				buf.WriteByte('\n')
				state = STATE_HUNK
			} else {
				// ignore other delta lines
				i++
			}
		case STATE_HUNK:
			buf.Write(line)
			i++
			state = STATE_LINE
		case STATE_LINE:
			if bytes.HasPrefix(line, []byte("diff")) {
				state = STATE_DELTA
			} else if bytes.HasPrefix(line, []byte("@@")) {
				state = STATE_HUNK
			} else if bytes.Contains([]byte("-+ "), line[:1]) {
				buf.Write(line)
				i++
			} else {
				i++
			}
		}
	}

	return FormattedDiff(buf.String())
}

// ForEachDelta iterates over the file deltas in a diff in a zero-copy manner
func (d FormattedDiff) ForEachDelta(f func(Delta) bool) {
	remaining := d
	var loc protocol.Location
	for len(remaining) > 0 {
		delta := scanDelta(string(remaining))
		remaining = remaining[len(delta):]

		newlineIdx := strings.IndexByte(delta, '\n')
		fileNameLine := delta[:newlineIdx]
		hunks := delta[newlineIdx+1:]
		oldFile, newFile := splitFileNames(fileNameLine)

		if cont := f(Delta{
			location: loc,
			oldFile:  oldFile,
			newFile:  newFile,
			hunks:    hunks,
		}); !cont {
			return
		}

		loc = loc.Shift(protocol.Location{
			Offset: len(delta),
			Line:   strings.Count(delta, "\n"),
		})
	}
}

func splitFileNames(fileNames string) (string, string) {
	// If there are no extra spaces, just split by space
	spaceSplit := strings.Split(fileNames, fileSeparator)
	if len(spaceSplit) == 2 {
		return spaceSplit[0], spaceSplit[1]
	}

	// If there are extra spaces, look for file endings that match
	lastChunk := spaceSplit[len(spaceSplit)-1]
	for i, chunk := range spaceSplit[:len(spaceSplit)-1] {
		if chunk == lastChunk {
			return strings.Join(spaceSplit[:i+1], fileSeparator), strings.Join(spaceSplit[i+1:], fileSeparator)
		}
	}

	// Otherwise, arbitrarily choose half the chunks for one file, and half for the other
	return strings.Join(spaceSplit[:len(spaceSplit)/2], fileSeparator),
		strings.Join(spaceSplit[len(spaceSplit)/2:], fileSeparator)
}

func scanDelta(s string) string {
	offset := 0
	for {
		idx := strings.IndexByte(s[offset:], '\n')
		if idx == -1 {
			return s
		}

		if idx+offset+1 == len(s) {
			return s
		}

		if strings.IndexByte("@+- <>=", s[idx+offset+1]) >= 0 {
			offset += idx + 1
		} else {
			return s[:offset+idx+1]
		}
	}
}

type Delta struct {
	location protocol.Location
	oldFile  string
	newFile  string
	hunks    string
}

func (d Delta) OldFile() (string, protocol.Location) {
	return d.oldFile, d.location
}

func (d Delta) NewFile() (string, protocol.Location) {
	return d.newFile, d.location.Shift(protocol.Location{
		Offset: len(d.newFile) + len(fileSeparator),
		Column: len(d.newFile) + len(fileSeparator),
	})
}

// ForEachHunk iterates over each hunk in a delta in a zero-copy manner
func (d Delta) ForEachHunk(f func(Hunk) bool) {
	remaining := d.hunks
	loc := d.location.Shift(protocol.Location{Line: 1, Offset: len(d.oldFile) + len(d.newFile) + len(fileSeparator) + len("\n")})
	for len(remaining) > 0 {
		hunk := scanHunk(remaining)
		remaining = remaining[len(hunk):]

		newlineIdx := strings.IndexByte(hunk, '\n')
		header := hunk[:newlineIdx]
		lines := hunk[newlineIdx+1:]

		if cont := f(Hunk{
			location: loc,
			header:   header,
			lines:    lines,
		}); !cont {
			return
		}

		loc = loc.Shift(protocol.Location{
			Offset: len(hunk),
			Line:   strings.Count(hunk, "\n"),
		})
	}
}

func scanHunk(s string) string {
	offset := 0
	for {
		idx := strings.IndexByte(s[offset:], '\n')
		if idx == -1 {
			return s
		}

		if idx+offset+1 == len(s) {
			return s
		}

		switch s[idx+offset+1] {
		case '@':
			return s[:offset+idx+1]
		}
		offset += idx + 1
	}
}

type Hunk struct {
	location protocol.Location
	header   string
	lines    string
}

// Header returns the @@-prefixed header for the hunk
func (h Hunk) Header() (string, protocol.Location) {
	return h.header, h.location
}

// ForEachLine iterates over each line in a hunk in a zero-copy manner
func (h Hunk) ForEachLine(f func(Line) bool) {
	remaining := h.lines
	loc := h.location.Shift(protocol.Location{Line: 1, Offset: len(h.header) + len("\n")})
	for len(remaining) > 0 {
		line := scanLine(remaining)
		remaining = remaining[len(line):]

		if cont := f(Line{
			location: loc,
			fullLine: line,
		}); !cont {
			return
		}

		loc = loc.Shift(protocol.Location{
			Offset: len(line),
			Line:   1,
		})
	}
}

func scanLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx > 0 {
		return s[:idx+1]
	}
	return s
}

type Line struct {
	location protocol.Location
	fullLine string
}

// Origin returns the first character of the line, which should be either '+', '-', or ' '
func (l Line) Origin() byte {
	return l.fullLine[0]
}

// Content returns the full content of the line, including the trailing newline
func (l Line) Content() (string, protocol.Location) {
	return l.fullLine[1:], l.location.Shift(protocol.Location{Column: 1, Offset: 1})
}

// DiffFetcher is a handle to the stdin and stdout of a git diff-tree subprocess
// started with StartDiffFetcher
type DiffFetcher struct {
	stdin   io.Writer
	stderr  bytes.Buffer
	scanner *bufio.Scanner
}

// StartDiffFetcher starts a git diff-tree subprocess that waits, listening on stdin
// for comimt hashes to generate patches for.
func StartDiffFetcher(ctx context.Context, dir string) (*DiffFetcher, error) {
	cmd := exec.CommandContext(ctx, "git", "diff-tree", "--stdin", "-p", "--format=format:")
	cmd.Dir = dir

	stdoutReader, stdoutWriter := io.Pipe()
	cmd.Stdout = stdoutWriter

	stdinReader, stdinWriter := io.Pipe()
	cmd.Stdin = stdinReader

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdoutReader)
	scanner.Buffer(make([]byte, 1024), 1<<30)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Note that this only works when we write to stdin, then read from stdout before writing
		// anything else to stdin, since we are using `HasSuffix` and not `Contains`.
		if bytes.HasSuffix(data, []byte("ENDOFPATCH\n")) {
			if bytes.Equal(data, []byte("ENDOFPATCH\n")) {
				// Empty patch
				return len(data), data[:0], nil
			}
			return len(data), data[:len(data)-len("ENDOFPATCH\n")], nil
		}

		return 0, nil, nil
	})

	return &DiffFetcher{
		stdin:   stdinWriter,
		scanner: scanner,
		stderr:  stderrBuf,
	}, nil
}

// FetchDiff fetches a diff from the git diff-tree subprocess, writing to its stdin
// and waiting for its response on stdout. Note that this is not safe to call concurrently.
func (d *DiffFetcher) FetchDiff(hash []byte) ([]byte, error) {
	// HACK: There is no way (as far as I can tell) to make `git diff-tree --stdin` to
	// write a trailing null byte or tell us how much to read in advance, and since we're
	// using a long-running process, the stream doesn't close at the end, and we can't use the
	// start of a new patch to signify end of patch since we want to be able to do each round-trip
	// serially. We resort to sending the subprocess a bogus commit hash named "EOF", which it
	// will fail to read as a tree, and print back to stdout literally. We use this as a signal
	// that the subprocess is done outputting for this commit.
	d.stdin.Write(append(hash, []byte("\nENDOFPATCH\n")...))

	if d.scanner.Scan() {
		return d.scanner.Bytes(), nil
	} else if err := d.scanner.Err(); err != nil {
		return nil, err
	} else if d.stderr.String() != "" {
		return nil, errors.Errorf("git subprocess stderr: %s", d.stderr.String())
	}
	return nil, errors.New("expected scan to succeed")
}
