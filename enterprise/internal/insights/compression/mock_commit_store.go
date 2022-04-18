// Code generated by go-mockgen 1.1.5; DO NOT EDIT.

package compression

import (
	"context"
	"sync"
	"time"

	api "github.com/sourcegraph/sourcegraph/internal/api"
	gitdomain "github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// MockCommitStore is a mock implementation of the CommitStore interface
// (from the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression)
// used for unit testing.
type MockCommitStore struct {
	// GetFunc is an instance of a mock function object controlling the
	// behavior of the method Get.
	GetFunc *CommitStoreGetFunc
	// GetMetadataFunc is an instance of a mock function object controlling
	// the behavior of the method GetMetadata.
	GetMetadataFunc *CommitStoreGetMetadataFunc
	// InsertCommitsFunc is an instance of a mock function object
	// controlling the behavior of the method InsertCommits.
	InsertCommitsFunc *CommitStoreInsertCommitsFunc
	// SaveFunc is an instance of a mock function object controlling the
	// behavior of the method Save.
	SaveFunc *CommitStoreSaveFunc
	// UpsertMetadataStampFunc is an instance of a mock function object
	// controlling the behavior of the method UpsertMetadataStamp.
	UpsertMetadataStampFunc *CommitStoreUpsertMetadataStampFunc
}

// NewMockCommitStore creates a new mock of the CommitStore interface. All
// methods return zero values for all results, unless overwritten.
func NewMockCommitStore() *MockCommitStore {
	return &MockCommitStore{
		GetFunc: &CommitStoreGetFunc{
			defaultHook: func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error) {
				return nil, nil
			},
		},
		GetMetadataFunc: &CommitStoreGetMetadataFunc{
			defaultHook: func(context.Context, api.RepoID) (CommitIndexMetadata, error) {
				return CommitIndexMetadata{}, nil
			},
		},
		InsertCommitsFunc: &CommitStoreInsertCommitsFunc{
			defaultHook: func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error {
				return nil
			},
		},
		SaveFunc: &CommitStoreSaveFunc{
			defaultHook: func(context.Context, api.RepoID, *gitdomain.Commit, string) error {
				return nil
			},
		},
		UpsertMetadataStampFunc: &CommitStoreUpsertMetadataStampFunc{
			defaultHook: func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error) {
				return CommitIndexMetadata{}, nil
			},
		},
	}
}

// NewStrictMockCommitStore creates a new mock of the CommitStore interface.
// All methods panic on invocation, unless overwritten.
func NewStrictMockCommitStore() *MockCommitStore {
	return &MockCommitStore{
		GetFunc: &CommitStoreGetFunc{
			defaultHook: func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error) {
				panic("unexpected invocation of MockCommitStore.Get")
			},
		},
		GetMetadataFunc: &CommitStoreGetMetadataFunc{
			defaultHook: func(context.Context, api.RepoID) (CommitIndexMetadata, error) {
				panic("unexpected invocation of MockCommitStore.GetMetadata")
			},
		},
		InsertCommitsFunc: &CommitStoreInsertCommitsFunc{
			defaultHook: func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error {
				panic("unexpected invocation of MockCommitStore.InsertCommits")
			},
		},
		SaveFunc: &CommitStoreSaveFunc{
			defaultHook: func(context.Context, api.RepoID, *gitdomain.Commit, string) error {
				panic("unexpected invocation of MockCommitStore.Save")
			},
		},
		UpsertMetadataStampFunc: &CommitStoreUpsertMetadataStampFunc{
			defaultHook: func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error) {
				panic("unexpected invocation of MockCommitStore.UpsertMetadataStamp")
			},
		},
	}
}

// NewMockCommitStoreFrom creates a new mock of the MockCommitStore
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockCommitStoreFrom(i CommitStore) *MockCommitStore {
	return &MockCommitStore{
		GetFunc: &CommitStoreGetFunc{
			defaultHook: i.Get,
		},
		GetMetadataFunc: &CommitStoreGetMetadataFunc{
			defaultHook: i.GetMetadata,
		},
		InsertCommitsFunc: &CommitStoreInsertCommitsFunc{
			defaultHook: i.InsertCommits,
		},
		SaveFunc: &CommitStoreSaveFunc{
			defaultHook: i.Save,
		},
		UpsertMetadataStampFunc: &CommitStoreUpsertMetadataStampFunc{
			defaultHook: i.UpsertMetadataStamp,
		},
	}
}

// CommitStoreGetFunc describes the behavior when the Get method of the
// parent MockCommitStore instance is invoked.
type CommitStoreGetFunc struct {
	defaultHook func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error)
	hooks       []func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error)
	history     []CommitStoreGetFuncCall
	mutex       sync.Mutex
}

// Get delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCommitStore) Get(v0 context.Context, v1 api.RepoID, v2 time.Time, v3 time.Time) ([]CommitStamp, error) {
	r0, r1 := m.GetFunc.nextHook()(v0, v1, v2, v3)
	m.GetFunc.appendCall(CommitStoreGetFuncCall{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Get method of the
// parent MockCommitStore instance is invoked and the hook queue is empty.
func (f *CommitStoreGetFunc) SetDefaultHook(hook func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Get method of the parent MockCommitStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *CommitStoreGetFunc) PushHook(hook func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *CommitStoreGetFunc) SetDefaultReturn(r0 []CommitStamp, r1 error) {
	f.SetDefaultHook(func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *CommitStoreGetFunc) PushReturn(r0 []CommitStamp, r1 error) {
	f.PushHook(func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error) {
		return r0, r1
	})
}

func (f *CommitStoreGetFunc) nextHook() func(context.Context, api.RepoID, time.Time, time.Time) ([]CommitStamp, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommitStoreGetFunc) appendCall(r0 CommitStoreGetFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommitStoreGetFuncCall objects describing
// the invocations of this function.
func (f *CommitStoreGetFunc) History() []CommitStoreGetFuncCall {
	f.mutex.Lock()
	history := make([]CommitStoreGetFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommitStoreGetFuncCall is an object that describes an invocation of
// method Get on an instance of MockCommitStore.
type CommitStoreGetFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 api.RepoID
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 time.Time
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 time.Time
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []CommitStamp
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommitStoreGetFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommitStoreGetFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// CommitStoreGetMetadataFunc describes the behavior when the GetMetadata
// method of the parent MockCommitStore instance is invoked.
type CommitStoreGetMetadataFunc struct {
	defaultHook func(context.Context, api.RepoID) (CommitIndexMetadata, error)
	hooks       []func(context.Context, api.RepoID) (CommitIndexMetadata, error)
	history     []CommitStoreGetMetadataFuncCall
	mutex       sync.Mutex
}

// GetMetadata delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCommitStore) GetMetadata(v0 context.Context, v1 api.RepoID) (CommitIndexMetadata, error) {
	r0, r1 := m.GetMetadataFunc.nextHook()(v0, v1)
	m.GetMetadataFunc.appendCall(CommitStoreGetMetadataFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetMetadata method
// of the parent MockCommitStore instance is invoked and the hook queue is
// empty.
func (f *CommitStoreGetMetadataFunc) SetDefaultHook(hook func(context.Context, api.RepoID) (CommitIndexMetadata, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetMetadata method of the parent MockCommitStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CommitStoreGetMetadataFunc) PushHook(hook func(context.Context, api.RepoID) (CommitIndexMetadata, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *CommitStoreGetMetadataFunc) SetDefaultReturn(r0 CommitIndexMetadata, r1 error) {
	f.SetDefaultHook(func(context.Context, api.RepoID) (CommitIndexMetadata, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *CommitStoreGetMetadataFunc) PushReturn(r0 CommitIndexMetadata, r1 error) {
	f.PushHook(func(context.Context, api.RepoID) (CommitIndexMetadata, error) {
		return r0, r1
	})
}

func (f *CommitStoreGetMetadataFunc) nextHook() func(context.Context, api.RepoID) (CommitIndexMetadata, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommitStoreGetMetadataFunc) appendCall(r0 CommitStoreGetMetadataFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommitStoreGetMetadataFuncCall objects
// describing the invocations of this function.
func (f *CommitStoreGetMetadataFunc) History() []CommitStoreGetMetadataFuncCall {
	f.mutex.Lock()
	history := make([]CommitStoreGetMetadataFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommitStoreGetMetadataFuncCall is an object that describes an invocation
// of method GetMetadata on an instance of MockCommitStore.
type CommitStoreGetMetadataFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 api.RepoID
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 CommitIndexMetadata
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommitStoreGetMetadataFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommitStoreGetMetadataFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// CommitStoreInsertCommitsFunc describes the behavior when the
// InsertCommits method of the parent MockCommitStore instance is invoked.
type CommitStoreInsertCommitsFunc struct {
	defaultHook func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error
	hooks       []func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error
	history     []CommitStoreInsertCommitsFuncCall
	mutex       sync.Mutex
}

// InsertCommits delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCommitStore) InsertCommits(v0 context.Context, v1 api.RepoID, v2 []*gitdomain.Commit, v3 time.Time, v4 string) error {
	r0 := m.InsertCommitsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.InsertCommitsFunc.appendCall(CommitStoreInsertCommitsFuncCall{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefaultHook sets function that is called when the InsertCommits method
// of the parent MockCommitStore instance is invoked and the hook queue is
// empty.
func (f *CommitStoreInsertCommitsFunc) SetDefaultHook(hook func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// InsertCommits method of the parent MockCommitStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CommitStoreInsertCommitsFunc) PushHook(hook func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *CommitStoreInsertCommitsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *CommitStoreInsertCommitsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error {
		return r0
	})
}

func (f *CommitStoreInsertCommitsFunc) nextHook() func(context.Context, api.RepoID, []*gitdomain.Commit, time.Time, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommitStoreInsertCommitsFunc) appendCall(r0 CommitStoreInsertCommitsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommitStoreInsertCommitsFuncCall objects
// describing the invocations of this function.
func (f *CommitStoreInsertCommitsFunc) History() []CommitStoreInsertCommitsFuncCall {
	f.mutex.Lock()
	history := make([]CommitStoreInsertCommitsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommitStoreInsertCommitsFuncCall is an object that describes an
// invocation of method InsertCommits on an instance of MockCommitStore.
type CommitStoreInsertCommitsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 api.RepoID
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 []*gitdomain.Commit
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 time.Time
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommitStoreInsertCommitsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommitStoreInsertCommitsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// CommitStoreSaveFunc describes the behavior when the Save method of the
// parent MockCommitStore instance is invoked.
type CommitStoreSaveFunc struct {
	defaultHook func(context.Context, api.RepoID, *gitdomain.Commit, string) error
	hooks       []func(context.Context, api.RepoID, *gitdomain.Commit, string) error
	history     []CommitStoreSaveFuncCall
	mutex       sync.Mutex
}

// Save delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCommitStore) Save(v0 context.Context, v1 api.RepoID, v2 *gitdomain.Commit, v3 string) error {
	r0 := m.SaveFunc.nextHook()(v0, v1, v2, v3)
	m.SaveFunc.appendCall(CommitStoreSaveFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Save method of the
// parent MockCommitStore instance is invoked and the hook queue is empty.
func (f *CommitStoreSaveFunc) SetDefaultHook(hook func(context.Context, api.RepoID, *gitdomain.Commit, string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Save method of the parent MockCommitStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *CommitStoreSaveFunc) PushHook(hook func(context.Context, api.RepoID, *gitdomain.Commit, string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *CommitStoreSaveFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, api.RepoID, *gitdomain.Commit, string) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *CommitStoreSaveFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, api.RepoID, *gitdomain.Commit, string) error {
		return r0
	})
}

func (f *CommitStoreSaveFunc) nextHook() func(context.Context, api.RepoID, *gitdomain.Commit, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommitStoreSaveFunc) appendCall(r0 CommitStoreSaveFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommitStoreSaveFuncCall objects describing
// the invocations of this function.
func (f *CommitStoreSaveFunc) History() []CommitStoreSaveFuncCall {
	f.mutex.Lock()
	history := make([]CommitStoreSaveFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommitStoreSaveFuncCall is an object that describes an invocation of
// method Save on an instance of MockCommitStore.
type CommitStoreSaveFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 api.RepoID
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 *gitdomain.Commit
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommitStoreSaveFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommitStoreSaveFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// CommitStoreUpsertMetadataStampFunc describes the behavior when the
// UpsertMetadataStamp method of the parent MockCommitStore instance is
// invoked.
type CommitStoreUpsertMetadataStampFunc struct {
	defaultHook func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error)
	hooks       []func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error)
	history     []CommitStoreUpsertMetadataStampFuncCall
	mutex       sync.Mutex
}

// UpsertMetadataStamp delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockCommitStore) UpsertMetadataStamp(v0 context.Context, v1 api.RepoID, v2 time.Time) (CommitIndexMetadata, error) {
	r0, r1 := m.UpsertMetadataStampFunc.nextHook()(v0, v1, v2)
	m.UpsertMetadataStampFunc.appendCall(CommitStoreUpsertMetadataStampFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the UpsertMetadataStamp
// method of the parent MockCommitStore instance is invoked and the hook
// queue is empty.
func (f *CommitStoreUpsertMetadataStampFunc) SetDefaultHook(hook func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// UpsertMetadataStamp method of the parent MockCommitStore instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *CommitStoreUpsertMetadataStampFunc) PushHook(hook func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *CommitStoreUpsertMetadataStampFunc) SetDefaultReturn(r0 CommitIndexMetadata, r1 error) {
	f.SetDefaultHook(func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *CommitStoreUpsertMetadataStampFunc) PushReturn(r0 CommitIndexMetadata, r1 error) {
	f.PushHook(func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error) {
		return r0, r1
	})
}

func (f *CommitStoreUpsertMetadataStampFunc) nextHook() func(context.Context, api.RepoID, time.Time) (CommitIndexMetadata, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommitStoreUpsertMetadataStampFunc) appendCall(r0 CommitStoreUpsertMetadataStampFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommitStoreUpsertMetadataStampFuncCall
// objects describing the invocations of this function.
func (f *CommitStoreUpsertMetadataStampFunc) History() []CommitStoreUpsertMetadataStampFuncCall {
	f.mutex.Lock()
	history := make([]CommitStoreUpsertMetadataStampFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommitStoreUpsertMetadataStampFuncCall is an object that describes an
// invocation of method UpsertMetadataStamp on an instance of
// MockCommitStore.
type CommitStoreUpsertMetadataStampFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 api.RepoID
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 time.Time
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 CommitIndexMetadata
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommitStoreUpsertMetadataStampFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommitStoreUpsertMetadataStampFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
