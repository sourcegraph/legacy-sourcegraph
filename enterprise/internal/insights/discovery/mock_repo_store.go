// Code generated by go-mockgen 1.2.0; DO NOT EDIT.

package discovery

import (
	"context"
	"sync"

	database "github.com/sourcegraph/sourcegraph/internal/database"
	types "github.com/sourcegraph/sourcegraph/internal/types"
)

// MockRepoStore is a mock implementation of the RepoStore interface (from
// the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery)
// used for unit testing.
type MockRepoStore struct {
	// ListFunc is an instance of a mock function object controlling the
	// behavior of the method List.
	ListFunc *RepoStoreListFunc
}

// NewMockRepoStore creates a new mock of the RepoStore interface. All
// methods return zero values for all results, unless overwritten.
func NewMockRepoStore() *MockRepoStore {
	return &MockRepoStore{
		ListFunc: &RepoStoreListFunc{
			defaultHook: func(context.Context, database.ReposListOptions) (r0 []*types.Repo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockRepoStore creates a new mock of the RepoStore interface. All
// methods panic on invocation, unless overwritten.
func NewStrictMockRepoStore() *MockRepoStore {
	return &MockRepoStore{
		ListFunc: &RepoStoreListFunc{
			defaultHook: func(context.Context, database.ReposListOptions) ([]*types.Repo, error) {
				panic("unexpected invocation of MockRepoStore.List")
			},
		},
	}
}

// NewMockRepoStoreFrom creates a new mock of the MockRepoStore interface.
// All methods delegate to the given implementation, unless overwritten.
func NewMockRepoStoreFrom(i RepoStore) *MockRepoStore {
	return &MockRepoStore{
		ListFunc: &RepoStoreListFunc{
			defaultHook: i.List,
		},
	}
}

// RepoStoreListFunc describes the behavior when the List method of the
// parent MockRepoStore instance is invoked.
type RepoStoreListFunc struct {
	defaultHook func(context.Context, database.ReposListOptions) ([]*types.Repo, error)
	hooks       []func(context.Context, database.ReposListOptions) ([]*types.Repo, error)
	history     []RepoStoreListFuncCall
	mutex       sync.Mutex
}

// List delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockRepoStore) List(v0 context.Context, v1 database.ReposListOptions) ([]*types.Repo, error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.appendCall(RepoStoreListFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the List method of the
// parent MockRepoStore instance is invoked and the hook queue is empty.
func (f *RepoStoreListFunc) SetDefaultHook(hook func(context.Context, database.ReposListOptions) ([]*types.Repo, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// List method of the parent MockRepoStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *RepoStoreListFunc) PushHook(hook func(context.Context, database.ReposListOptions) ([]*types.Repo, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *RepoStoreListFunc) SetDefaultReturn(r0 []*types.Repo, r1 error) {
	f.SetDefaultHook(func(context.Context, database.ReposListOptions) ([]*types.Repo, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *RepoStoreListFunc) PushReturn(r0 []*types.Repo, r1 error) {
	f.PushHook(func(context.Context, database.ReposListOptions) ([]*types.Repo, error) {
		return r0, r1
	})
}

func (f *RepoStoreListFunc) nextHook() func(context.Context, database.ReposListOptions) ([]*types.Repo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoStoreListFunc) appendCall(r0 RepoStoreListFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of RepoStoreListFuncCall objects describing
// the invocations of this function.
func (f *RepoStoreListFunc) History() []RepoStoreListFuncCall {
	f.mutex.Lock()
	history := make([]RepoStoreListFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoStoreListFuncCall is an object that describes an invocation of method
// List on an instance of MockRepoStore.
type RepoStoreListFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 database.ReposListOptions
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []*types.Repo
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c RepoStoreListFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c RepoStoreListFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
