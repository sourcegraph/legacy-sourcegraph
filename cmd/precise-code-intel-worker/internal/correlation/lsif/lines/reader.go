package lines

import (
	"bufio"
	"context"
	"io"
	"runtime"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

// LineBufferSize is the maximum size of the buffer used to read each line of a raw LSIF index. Lines in
// LSIF can get very long as it include escaped hover text (package documentation), as well as large edges
// such as the contains edge of large documents.
//
// This corresponds a 10MB buffer that can accommodate 10 million characters.
const LineBufferSize = 1e7

// ChannelBufferSize is the number sources lines that can be read ahead of the correlator.
const ChannelBufferSize = 512

// NumUnmarshalGoRoutines is the number of goroutines launched to unmarshal individual lines.
var NumUnmarshalGoRoutines = runtime.NumCPU()

// Read reads the given content as line-separated objects which are unmarshallable by the given function
// and returns a channel of lsif.Pair values for each non-empty line.
func Read(ctx context.Context, r io.Reader, unmarshal func(line []byte) (_ lsif.Element, err error)) <-chan lsif.Pair {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, LineBufferSize), LineBufferSize)

	// Read the document in a separate go-routine.
	lineCh := make(chan []byte, ChannelBufferSize)
	go func() {
		defer close(lineCh)

		for scanner.Scan() {
			if line := scanner.Bytes(); len(line) != 0 {
				lineCh <- line
			}
		}
	}()

	pairCh := make(chan lsif.Pair, ChannelBufferSize)
	go func() {
		defer close(pairCh)

		// Unmarshal workers receive work assignments as indices into a shared
		// slice and put the result into the same index in a second shared slice.
		work := make(chan int, NumUnmarshalGoRoutines)
		defer close(work)

		// Each unmarshal worker sends a zero-length value on this channel
		// to signal completion of a unit of work.
		signal := make(chan struct{}, NumUnmarshalGoRoutines)
		defer close(signal)

		// The input slice
		lines := make([][]byte, NumUnmarshalGoRoutines)

		// The result slice
		pairs := make([]lsif.Pair, NumUnmarshalGoRoutines)

		for i := 0; i < NumUnmarshalGoRoutines; i++ {
			go func() {
				for idx := range work {
					element, err := unmarshal(lines[idx])
					pairs[idx].Element = element
					pairs[idx].Err = err
					signal <- struct{}{}
				}
			}()
		}

		done := false
		for !done {
			i := 0

			// Read a new "batch" of lines from the reader routine and fill the
			// shared array. Each index that receives a new value is queued in
			// the unmarshal worker channel and can be immediately processed.
			for i < NumUnmarshalGoRoutines {
				line, ok := <-lineCh
				if !ok {
					done = true
					break
				}

				lines[i] = line
				work <- i
				i++
			}

			// Wait until the current batch has bee completely unmarshalled
			for j := 0; j < i; j++ {
				<-signal
			}

			// Read the result array in order. If the caller context has completed,
			// we'll abandon any additional values we were going to send on this
			// channel (as well as any additional errors from the scanner).
			for j := 0; j < i; j++ {
				select {
				case pairCh <- pairs[j]:
				case <-ctx.Done():
					return
				}
			}
		}

		// If there was an error reading from the source, output it here
		if err := scanner.Err(); err != nil {
			pairCh <- lsif.Pair{Err: err}
		}
	}()

	return pairCh
}
