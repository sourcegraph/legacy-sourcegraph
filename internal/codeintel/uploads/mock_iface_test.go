// Code generated by go-mockgen 1.3.1; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the metadata.yaml file in the root of this repository.

package uploads

import (
	"context"
	"sync"
	"time"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	shared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

// MockStore is a mock implementation of the Store interface (from the
// package
// github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store)
// used for unit testing.
type MockStore struct {
	// DeleteIndexesWithoutRepositoryFunc is an instance of a mock function
	// object controlling the behavior of the method
	// DeleteIndexesWithoutRepository.
	DeleteIndexesWithoutRepositoryFunc *StoreDeleteIndexesWithoutRepositoryFunc
	// DeleteUploadsWithoutRepositoryFunc is an instance of a mock function
	// object controlling the behavior of the method
	// DeleteUploadsWithoutRepository.
	DeleteUploadsWithoutRepositoryFunc *StoreDeleteUploadsWithoutRepositoryFunc
	// ListFunc is an instance of a mock function object controlling the
	// behavior of the method List.
	ListFunc *StoreListFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		DeleteIndexesWithoutRepositoryFunc: &StoreDeleteIndexesWithoutRepositoryFunc{
			defaultHook: func(context.Context, time.Time) (r0 map[int]int, r1 error) {
				return
			},
		},
		DeleteUploadsWithoutRepositoryFunc: &StoreDeleteUploadsWithoutRepositoryFunc{
			defaultHook: func(context.Context, time.Time) (r0 map[int]int, r1 error) {
				return
			},
		},
		ListFunc: &StoreListFunc{
			defaultHook: func(context.Context, store.ListOpts) (r0 []shared.Upload, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore creates a new mock of the Store interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		DeleteIndexesWithoutRepositoryFunc: &StoreDeleteIndexesWithoutRepositoryFunc{
			defaultHook: func(context.Context, time.Time) (map[int]int, error) {
				panic("unexpected invocation of MockStore.DeleteIndexesWithoutRepository")
			},
		},
		DeleteUploadsWithoutRepositoryFunc: &StoreDeleteUploadsWithoutRepositoryFunc{
			defaultHook: func(context.Context, time.Time) (map[int]int, error) {
				panic("unexpected invocation of MockStore.DeleteUploadsWithoutRepository")
			},
		},
		ListFunc: &StoreListFunc{
			defaultHook: func(context.Context, store.ListOpts) ([]shared.Upload, error) {
				panic("unexpected invocation of MockStore.List")
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i store.Store) *MockStore {
	return &MockStore{
		DeleteIndexesWithoutRepositoryFunc: &StoreDeleteIndexesWithoutRepositoryFunc{
			defaultHook: i.DeleteIndexesWithoutRepository,
		},
		DeleteUploadsWithoutRepositoryFunc: &StoreDeleteUploadsWithoutRepositoryFunc{
			defaultHook: i.DeleteUploadsWithoutRepository,
		},
		ListFunc: &StoreListFunc{
			defaultHook: i.List,
		},
	}
}

// StoreDeleteIndexesWithoutRepositoryFunc describes the behavior when the
// DeleteIndexesWithoutRepository method of the parent MockStore instance is
// invoked.
type StoreDeleteIndexesWithoutRepositoryFunc struct {
	defaultHook func(context.Context, time.Time) (map[int]int, error)
	hooks       []func(context.Context, time.Time) (map[int]int, error)
	history     []StoreDeleteIndexesWithoutRepositoryFuncCall
	mutex       sync.Mutex
}

// DeleteIndexesWithoutRepository delegates to the next hook function in the
// queue and stores the parameter and result values of this invocation.
func (m *MockStore) DeleteIndexesWithoutRepository(v0 context.Context, v1 time.Time) (map[int]int, error) {
	r0, r1 := m.DeleteIndexesWithoutRepositoryFunc.nextHook()(v0, v1)
	m.DeleteIndexesWithoutRepositoryFunc.appendCall(StoreDeleteIndexesWithoutRepositoryFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// DeleteIndexesWithoutRepository method of the parent MockStore instance is
// invoked and the hook queue is empty.
func (f *StoreDeleteIndexesWithoutRepositoryFunc) SetDefaultHook(hook func(context.Context, time.Time) (map[int]int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// DeleteIndexesWithoutRepository method of the parent MockStore instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *StoreDeleteIndexesWithoutRepositoryFunc) PushHook(hook func(context.Context, time.Time) (map[int]int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreDeleteIndexesWithoutRepositoryFunc) SetDefaultReturn(r0 map[int]int, r1 error) {
	f.SetDefaultHook(func(context.Context, time.Time) (map[int]int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreDeleteIndexesWithoutRepositoryFunc) PushReturn(r0 map[int]int, r1 error) {
	f.PushHook(func(context.Context, time.Time) (map[int]int, error) {
		return r0, r1
	})
}

func (f *StoreDeleteIndexesWithoutRepositoryFunc) nextHook() func(context.Context, time.Time) (map[int]int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteIndexesWithoutRepositoryFunc) appendCall(r0 StoreDeleteIndexesWithoutRepositoryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDeleteIndexesWithoutRepositoryFuncCall
// objects describing the invocations of this function.
func (f *StoreDeleteIndexesWithoutRepositoryFunc) History() []StoreDeleteIndexesWithoutRepositoryFuncCall {
	f.mutex.Lock()
	history := make([]StoreDeleteIndexesWithoutRepositoryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteIndexesWithoutRepositoryFuncCall is an object that describes
// an invocation of method DeleteIndexesWithoutRepository on an instance of
// MockStore.
type StoreDeleteIndexesWithoutRepositoryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 time.Time
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[int]int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDeleteIndexesWithoutRepositoryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDeleteIndexesWithoutRepositoryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreDeleteUploadsWithoutRepositoryFunc describes the behavior when the
// DeleteUploadsWithoutRepository method of the parent MockStore instance is
// invoked.
type StoreDeleteUploadsWithoutRepositoryFunc struct {
	defaultHook func(context.Context, time.Time) (map[int]int, error)
	hooks       []func(context.Context, time.Time) (map[int]int, error)
	history     []StoreDeleteUploadsWithoutRepositoryFuncCall
	mutex       sync.Mutex
}

// DeleteUploadsWithoutRepository delegates to the next hook function in the
// queue and stores the parameter and result values of this invocation.
func (m *MockStore) DeleteUploadsWithoutRepository(v0 context.Context, v1 time.Time) (map[int]int, error) {
	r0, r1 := m.DeleteUploadsWithoutRepositoryFunc.nextHook()(v0, v1)
	m.DeleteUploadsWithoutRepositoryFunc.appendCall(StoreDeleteUploadsWithoutRepositoryFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// DeleteUploadsWithoutRepository method of the parent MockStore instance is
// invoked and the hook queue is empty.
func (f *StoreDeleteUploadsWithoutRepositoryFunc) SetDefaultHook(hook func(context.Context, time.Time) (map[int]int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// DeleteUploadsWithoutRepository method of the parent MockStore instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *StoreDeleteUploadsWithoutRepositoryFunc) PushHook(hook func(context.Context, time.Time) (map[int]int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreDeleteUploadsWithoutRepositoryFunc) SetDefaultReturn(r0 map[int]int, r1 error) {
	f.SetDefaultHook(func(context.Context, time.Time) (map[int]int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreDeleteUploadsWithoutRepositoryFunc) PushReturn(r0 map[int]int, r1 error) {
	f.PushHook(func(context.Context, time.Time) (map[int]int, error) {
		return r0, r1
	})
}

func (f *StoreDeleteUploadsWithoutRepositoryFunc) nextHook() func(context.Context, time.Time) (map[int]int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteUploadsWithoutRepositoryFunc) appendCall(r0 StoreDeleteUploadsWithoutRepositoryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDeleteUploadsWithoutRepositoryFuncCall
// objects describing the invocations of this function.
func (f *StoreDeleteUploadsWithoutRepositoryFunc) History() []StoreDeleteUploadsWithoutRepositoryFuncCall {
	f.mutex.Lock()
	history := make([]StoreDeleteUploadsWithoutRepositoryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteUploadsWithoutRepositoryFuncCall is an object that describes
// an invocation of method DeleteUploadsWithoutRepository on an instance of
// MockStore.
type StoreDeleteUploadsWithoutRepositoryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 time.Time
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[int]int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDeleteUploadsWithoutRepositoryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDeleteUploadsWithoutRepositoryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreListFunc describes the behavior when the List method of the parent
// MockStore instance is invoked.
type StoreListFunc struct {
	defaultHook func(context.Context, store.ListOpts) ([]shared.Upload, error)
	hooks       []func(context.Context, store.ListOpts) ([]shared.Upload, error)
	history     []StoreListFuncCall
	mutex       sync.Mutex
}

// List delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) List(v0 context.Context, v1 store.ListOpts) ([]shared.Upload, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.appendCall(StoreListFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the List method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreListFunc) SetDefaultHook(hook func(context.Context, store.ListOpts) ([]shared.Upload, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// List method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreListFunc) PushHook(hook func(context.Context, store.ListOpts) ([]shared.Upload, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreListFunc) SetDefaultReturn(r0 []shared.Upload, r1 error) {
	f.SetDefaultHook(func(context.Context, store.ListOpts) ([]shared.Upload, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreListFunc) PushReturn(r0 []shared.Upload, r1 error) {
	f.PushHook(func(context.Context, store.ListOpts) ([]shared.Upload, error) {
		return r0, r1
	})
}

func (f *StoreListFunc) nextHook() func(context.Context, store.ListOpts) ([]shared.Upload, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreListFunc) appendCall(r0 StoreListFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreListFuncCall objects describing the
// invocations of this function.
func (f *StoreListFunc) History() []StoreListFuncCall {
	f.mutex.Lock()
	history := make([]StoreListFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreListFuncCall is an object that describes an invocation of method
// List on an instance of MockStore.
type StoreListFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.ListOpts
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []shared.Upload
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreListFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreListFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
