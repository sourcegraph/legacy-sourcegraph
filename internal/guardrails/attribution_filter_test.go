package guardrails_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/guardrails"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type fakeClient struct {
	mu sync.Mutex
	events []types.CompletionResponse
	err error
}

func (s *fakeClient) stream(e types.CompletionResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
	return s.err
}

func (s *fakeClient) trimmedDiffs() []string {
	var prefix string
	var diffs []string
	for _, e := range s.events {
		diffs = append(diffs, strings.TrimSpace(strings.TrimPrefix(e.Completion, prefix)))
		prefix = e.Completion
	}
	return diffs
}

type fakeSearch struct {
	mu sync.Mutex
	snippets []string
	response chan bool
}

func (s *fakeSearch) test(_ context.Context, snippet string) (bool, error) {
	s.mu.Lock()
	s.snippets = append(s.snippets, snippet)
	s.mu.Unlock()
	return <-s.response, nil
}

type eventOrder []event

func (o eventOrder) replay(ctx context.Context, f *guardrails.CompletionsFilter) error {
	var completionPrefix string
	for _, e := range o {
		if s := e.NextCompletionLine(); s != nil {
			completionPrefix = fmt.Sprintf("%s\n%s", completionPrefix, *s)
			if err := f.Send(ctx, types.CompletionResponse{
				Completion: completionPrefix,
			}); err != nil {
				return err
			}
		}
		e.Run()
	}
	return nil
}

type event interface {
	Run()
	NextCompletionLine() *string
}

type nextLine string
func (_ nextLine) Run() {}
func (n nextLine) NextCompletionLine() *string { s := string(n); return &s }

type searchFinishes struct { search *fakeSearch; canUseSnippet bool }
func (f searchFinishes) Run() { f.search.response <- f.canUseSnippet }
func (_ searchFinishes) NextCompletionLine() *string { return nil }

type contextCancelled func ()
func (c contextCancelled) Run() { c() }
func (_ contextCancelled) NextCompletionLine() *string { return nil }

func TestAttributionNotFound(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		searchFinishes{search: search, canUseSnippet: true},
	}
	require.NoError(t, o.replay(context.Background(), f))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Completion with lines 9 and 10 came together after
		// attribution search finished
		"9\n10",
	}
	require.Equal(t, want, got)
}

func TestAttributionFound(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		searchFinishes{search: search, canUseSnippet: false},
	}
	require.NoError(t, o.replay(context.Background(), f))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Completion with lines 9 and 10 never arrives,
		// as attribution was found
		// "9\n10",
	}
	require.Equal(t, want, got)
}

func TestAttributionNotFoundMoreDataAfter(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		searchFinishes{search: search, canUseSnippet: true},
		nextLine("11"),
		nextLine("12"),
	}
	require.NoError(t, o.replay(context.Background(), f))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Completion with lines 9 and 10 came together after
		// attribution search finished
		"9\n10",
		// Lines 11 and 12 came after search finished, they
		// are streamed through.
		"11", "12",
	}
	require.Equal(t, want, got)
}

func TestAttributionFoundMoreDataAfter(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		searchFinishes{search: search, canUseSnippet: false},
		nextLine("11"),
		nextLine("12"),
	}
	require.NoError(t, o.replay(context.Background(), f))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// No lines beyond 8 comve since attribution search
		// disallowed it:
		// "9\n10", "11", "12"
	}
	require.Equal(t, want, got)
}

func TestTimeout(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		contextCancelled(cancel),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
	}
	require.NoError(t, o.replay(ctx, f))
	require.NoError(t, f.WaitDone(ctx))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5",
		// Request cancelled before the rest of events arrived:
		// "6", "7", "8", "9", "10",
	}
	require.Equal(t, want, got)
}

func TestTimeoutAfterAttributionFound(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		searchFinishes{search: search, canUseSnippet: true},
		nextLine("11"),
		contextCancelled(cancel),
		nextLine("12"),
	}
	require.NoError(t, o.replay(ctx, f))
	require.NoError(t, f.WaitDone(ctx))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Completion with lines 9 and 10 arrive together,
		// as attribution was found
		"9\n10",
		// Line 11 manages to arrive while request finishes.
		"11",
		// Timeout. Line 12 never arrives:
		// "12",
	}
	require.Equal(t, want, got)
}

func TestTimeoutBeforeAttributionFound(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		contextCancelled(cancel),
		nextLine("10"),
		searchFinishes{search: search, canUseSnippet: false},
		nextLine("11"),
	}
	require.NoError(t, o.replay(ctx, f))
	require.NoError(t, f.WaitDone(ctx))
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Completion with lines 9 and 10 never arrives,
		// because attribution response arrives only after
		// time runs out. Same with the subsequent line.
		// "9\n10", "11"
	}
	require.Equal(t, want, got)
}

func TestAttributionSearchFinishesAfterWaitDoneIsCalled(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	ctx := context.Background()
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		nextLine("11"),
	}
	require.NoError(t, o.replay(ctx, f))
	var wg sync.WaitGroup
	wg.Add(1)
	go func () {
		require.NoError(t, f.WaitDone(ctx))
		wg.Done()
	}()
	got := client.trimmedDiffs()
	want := []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Lines that came over while attribution runs
		// not streamed yet
	}
	require.Equal(t, want, got)
	search.response <- true // Finish attribution search.
	wg.Wait() // WaitDone returns.
	got = client.trimmedDiffs()
	want = []string{
		"1", "2", "3", "4", "5", "6", "7", "8",
		// Lines that came over while attribution runs
		// not streamed as part of WaitDone.
		"9\n10\n11",
	}
	require.Equal(t, want, got)
}

func TestWaitDoneErr(t *testing.T) {
	client := &fakeClient{}
	search := &fakeSearch{response: make(chan bool)}
	f, err := guardrails.NewCompletionsFilter(guardrails.CompletionsFilterConfig{
		Sink: client.stream,
		Test: search.test,
		AttributionError: func (error) {},
	})
	require.NoError(t, err)
	ctx := context.Background()
	o := eventOrder{
		nextLine("1"),
		nextLine("2"),
		nextLine("3"),
		nextLine("4"),
		nextLine("5"),
		nextLine("6"),
		nextLine("7"),
		nextLine("8"),
		nextLine("9"),
		nextLine("10"),
		nextLine("11"),
	}
	require.NoError(t, o.replay(ctx, f))
	want := errors.New("fake!")
	gotErr := make(chan error)
	go func () {
		gotErr <- f.WaitDone(ctx)
	}()
	client.err = want// Client's send in WaitDone will err.
	search.response <- true // Finish attribution search.
	got := <-gotErr
	require.ErrorIs(t, got, want)
}
