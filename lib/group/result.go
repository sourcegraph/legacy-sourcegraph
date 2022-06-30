package group

import (
	"context"
	"sync"
)

// NewWithResults creates a new group that aggregates the return values
// of the functions passed to its `Go` method.
func NewWithResults[T any]() ResultGroup[T] {
	return &resultGroup[T]{
		Group: New(),
	}
}

// ResultGroup is a group that runs tasks that return a value.
type ResultGroup[T any] interface {
	// Go starts a task in a goroutine and collects its result. It will
	// not return until the goroutine has been started.
	Go(func() T)

	// Wait blocks until all goroutines started with Go() have completed.
	// It returns the collection of return values from the started tasks. There
	// are no guarantees about the order of the slice.
	Wait() []T

	// Configuration methods. See interface definitions for details.
	Contextable[ResultContextErrorGroup[T]]
	Errorable[ResultErrorGroup[T]]
	Limitable[ResultGroup[T]]
}

// ResultErrorGroup is a group that runs tasks that return a value and an error.
type ResultErrorGroup[T any] interface {
	// Go starts a task in a goroutine and collects its result. It will
	// not return until the goroutine has been started.
	Go(func() (T, error))

	// Wait blocks until all goroutines started with Go() have completed.
	// It returns the collection of return values from the started tasks. There
	// are no guarantees about the order of the slice. Additionally, it returns
	// a combined error composed of any non-nil errors returned from the tasks.
	Wait() ([]T, error)

	// WithCollectErrored configures the group to collect results even from
	// tasks that errored. By default, the return values from errored tasks are
	// dropped.
	WithCollectErrored() ResultErrorGroup[T]

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ResultErrorGroup[T]

	// Configuration methods. See interface definitions for details.
	Contextable[ResultContextErrorGroup[T]]
	Limitable[ResultErrorGroup[T]]
}

// ResultErrorGroup is a group that runs tasks that require a context and
// return a value and an error.
type ResultContextErrorGroup[T any] interface {
	// Go starts a task in a goroutine and collects its result. It will
	// not return until the goroutine has been started.
	Go(func(context.Context) (T, error))

	// Wait blocks until all goroutines started with Go() have completed.
	// It returns the collection of return values from the started tasks. There
	// are no guarantees about the order of the slice. Additionally, it returns
	// a combined error composed of any non-nil errors returned from the tasks.
	Wait() ([]T, error)

	// WithCollectErrored configures the group to collect results even from
	// tasks that errored. By default, the return values from errored tasks are
	// dropped.
	WithCollectErrored() ResultContextErrorGroup[T]

	// WithCancelOnError will cancel the group's context whenever any of the
	// functions started with Go() return an error.
	WithCancelOnError() ResultContextErrorGroup[T]

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ResultContextErrorGroup[T]

	// Configuration methods. See interface definitions for details.
	Limitable[ResultContextErrorGroup[T]]
}

type resultAggregator[T any] struct {
	mu      sync.Mutex
	results []T
}

func (r *resultAggregator[T]) add(res T) {
	r.mu.Lock()
	r.results = append(r.results, res)
	r.mu.Unlock()
}

type resultGroup[T any] struct {
	Group
	resultAggregator[T]
}

func (g *resultGroup[T]) Go(f func() T) {
	g.Group.Go(func() {
		res := f()
		g.add(res)
	})
}

func (g *resultGroup[T]) Wait() []T {
	g.Group.Wait()
	return g.results
}

func (g *resultGroup[T]) WithErrors() ResultErrorGroup[T] {
	return &resultErrorGroup[T]{
		ErrorGroup: g.Group.WithErrors(),
	}
}

func (g *resultGroup[T]) WithContext(ctx context.Context) ResultContextErrorGroup[T] {
	return &resultContextErrorGroup[T]{
		ContextErrorGroup: g.Group.WithContext(ctx),
	}
}

func (g *resultGroup[T]) WithLimit(limit int) ResultGroup[T] {
	g.Group = g.Group.WithLimit(limit)
	return g
}

func (g *resultGroup[T]) WithLimiter(limiter Limiter) ResultGroup[T] {
	g.Group = g.Group.WithLimiter(limiter)
	return g
}

type resultErrorGroup[T any] struct {
	ErrorGroup
	resultAggregator[T]
	collectErrored bool
}

func (g *resultErrorGroup[T]) Go(f func() (T, error)) {
	g.ErrorGroup.Go(func() error {
		res, err := f()
		if err == nil || g.collectErrored {
			g.add(res)
		}
		return err
	})
}

func (g *resultErrorGroup[T]) Wait() ([]T, error) {
	err := g.ErrorGroup.Wait()
	return g.results, err
}

func (g *resultErrorGroup[T]) WithCollectErrored() ResultErrorGroup[T] {
	g.collectErrored = true
	return g
}

func (g *resultErrorGroup[T]) WithFirstError() ResultErrorGroup[T] {
	g.ErrorGroup = g.ErrorGroup.WithFirstError()
	return g
}

func (g *resultErrorGroup[T]) WithLimit(limit int) ResultErrorGroup[T] {
	g.ErrorGroup = g.ErrorGroup.WithLimit(limit)
	return g
}

func (g *resultErrorGroup[T]) WithLimiter(limiter Limiter) ResultErrorGroup[T] {
	g.ErrorGroup = g.ErrorGroup.WithLimiter(limiter)
	return g
}

func (g *resultErrorGroup[T]) WithContext(ctx context.Context) ResultContextErrorGroup[T] {
	return &resultContextErrorGroup[T]{
		ContextErrorGroup: g.ErrorGroup.WithContext(ctx),
	}
}

type resultContextErrorGroup[T any] struct {
	ContextErrorGroup
	resultAggregator[T]
	collectErrored bool
}

func (g *resultContextErrorGroup[T]) Go(f func(context.Context) (T, error)) {
	g.ContextErrorGroup.Go(func(ctx context.Context) error {
		res, err := f(ctx)
		if err == nil || g.collectErrored {
			g.add(res)
		}
		return err
	})
}

func (g *resultContextErrorGroup[T]) Wait() ([]T, error) {
	err := g.ContextErrorGroup.Wait()
	return g.results, err
}

func (g *resultContextErrorGroup[T]) WithCollectErrored() ResultContextErrorGroup[T] {
	g.collectErrored = true
	return g
}

func (g *resultContextErrorGroup[T]) WithLimit(limit int) ResultContextErrorGroup[T] {
	g.ContextErrorGroup = g.ContextErrorGroup.WithLimit(limit)
	return g
}

func (g *resultContextErrorGroup[T]) WithLimiter(limiter Limiter) ResultContextErrorGroup[T] {
	g.ContextErrorGroup = g.ContextErrorGroup.WithLimiter(limiter)
	return g
}

func (g *resultContextErrorGroup[T]) WithCancelOnError() ResultContextErrorGroup[T] {
	g.ContextErrorGroup = g.ContextErrorGroup.WithCancelOnError()
	return g
}

func (g *resultContextErrorGroup[T]) WithFirstError() ResultContextErrorGroup[T] {
	g.ContextErrorGroup = g.ContextErrorGroup.WithFirstError()
	return g
}
