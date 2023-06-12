package common

import (
	"container/list"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type testJob struct {
	Value string
}

func (t *testJob) Identifier() string { return t.Value }

func (t *testJob) UUID() string { return t.Value }

func TestQueue(t *testing.T) {
	ctx := observation.NewContext(logtest.NoOp(t))
	queue := NewQueue[*testJob](ctx, "test-foo", list.New())

	if queue.name != "test_foo" {
		t.Fatalf("Expected queue name test_foo, but got %s", queue.name)
	}

	if !queue.Empty() {
		t.Error("Expected queue to be empty initially")
	}

	jobs := []testJob{
		{Value: "1"},
		{Value: "2"},
		{Value: "3"},
	}

	// Push 1, 2 and 3 into the queue.
	for _, j := range jobs {
		j := j
		queue.Push(&j)
	}

	if queue.Empty() {
		t.Error("Expected queue to not be empty after pushing elements")
	}

	// Pop and expect 1, 2 and 3 in that order (FIFO queue).
	for _, j := range jobs {
		expected := j
		gotJob := queue.Pop()

		if diff := cmp.Diff(expected, **gotJob); diff != "" {
			t.Errorf("mismatch in job, (-want, +got)\n%s", diff)
		}
	}

	if !queue.Empty() {
		t.Error("Expected queue to be empty after popping all elements")
	}
}
