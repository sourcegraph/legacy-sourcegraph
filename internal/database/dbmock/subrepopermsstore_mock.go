// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package dbmock

import (
	"context"
	"sync"

	api "github.com/sourcegraph/sourcegraph/internal/api"
	authz "github.com/sourcegraph/sourcegraph/internal/authz"
	database "github.com/sourcegraph/sourcegraph/internal/database"
	basestore "github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// MockSubRepoPermsStore is a mock implementation of the SubRepoPermsStore
// interface (from the package
// github.com/sourcegraph/sourcegraph/internal/database) used for unit
// testing.
type MockSubRepoPermsStore struct {
	// DoneFunc is an instance of a mock function object controlling the
	// behavior of the method Done.
	DoneFunc *SubRepoPermsStoreDoneFunc
	// GetFunc is an instance of a mock function object controlling the
	// behavior of the method Get.
	GetFunc *SubRepoPermsStoreGetFunc
	// GetByUserFunc is an instance of a mock function object controlling
	// the behavior of the method GetByUser.
	GetByUserFunc *SubRepoPermsStoreGetByUserFunc
	// TransactFunc is an instance of a mock function object controlling the
	// behavior of the method Transact.
	TransactFunc *SubRepoPermsStoreTransactFunc
	// UpsertFunc is an instance of a mock function object controlling the
	// behavior of the method Upsert.
	UpsertFunc *SubRepoPermsStoreUpsertFunc
	// UpsertWithSpecFunc is an instance of a mock function object
	// controlling the behavior of the method UpsertWithSpec.
	UpsertWithSpecFunc *SubRepoPermsStoreUpsertWithSpecFunc
	// WithFunc is an instance of a mock function object controlling the
	// behavior of the method With.
	WithFunc *SubRepoPermsStoreWithFunc
}

// NewMockSubRepoPermsStore creates a new mock of the SubRepoPermsStore
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockSubRepoPermsStore() *MockSubRepoPermsStore {
	return &MockSubRepoPermsStore{
		DoneFunc: &SubRepoPermsStoreDoneFunc{
			defaultHook: func(error) error {
				return nil
			},
		},
		GetFunc: &SubRepoPermsStoreGetFunc{
			defaultHook: func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error) {
				return nil, nil
			},
		},
		GetByUserFunc: &SubRepoPermsStoreGetByUserFunc{
			defaultHook: func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
				return nil, nil
			},
		},
		TransactFunc: &SubRepoPermsStoreTransactFunc{
			defaultHook: func(context.Context) (database.SubRepoPermsStore, error) {
				return nil, nil
			},
		},
		UpsertFunc: &SubRepoPermsStoreUpsertFunc{
			defaultHook: func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error {
				return nil
			},
		},
		UpsertWithSpecFunc: &SubRepoPermsStoreUpsertWithSpecFunc{
			defaultHook: func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error {
				return nil
			},
		},
		WithFunc: &SubRepoPermsStoreWithFunc{
			defaultHook: func(basestore.ShareableStore) database.SubRepoPermsStore {
				return nil
			},
		},
	}
}

// NewStrictMockSubRepoPermsStore creates a new mock of the
// SubRepoPermsStore interface. All methods panic on invocation, unless
// overwritten.
func NewStrictMockSubRepoPermsStore() *MockSubRepoPermsStore {
	return &MockSubRepoPermsStore{
		DoneFunc: &SubRepoPermsStoreDoneFunc{
			defaultHook: func(error) error {
				panic("unexpected invocation of MockSubRepoPermsStore.Done")
			},
		},
		GetFunc: &SubRepoPermsStoreGetFunc{
			defaultHook: func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error) {
				panic("unexpected invocation of MockSubRepoPermsStore.Get")
			},
		},
		GetByUserFunc: &SubRepoPermsStoreGetByUserFunc{
			defaultHook: func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
				panic("unexpected invocation of MockSubRepoPermsStore.GetByUser")
			},
		},
		TransactFunc: &SubRepoPermsStoreTransactFunc{
			defaultHook: func(context.Context) (database.SubRepoPermsStore, error) {
				panic("unexpected invocation of MockSubRepoPermsStore.Transact")
			},
		},
		UpsertFunc: &SubRepoPermsStoreUpsertFunc{
			defaultHook: func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error {
				panic("unexpected invocation of MockSubRepoPermsStore.Upsert")
			},
		},
		UpsertWithSpecFunc: &SubRepoPermsStoreUpsertWithSpecFunc{
			defaultHook: func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error {
				panic("unexpected invocation of MockSubRepoPermsStore.UpsertWithSpec")
			},
		},
		WithFunc: &SubRepoPermsStoreWithFunc{
			defaultHook: func(basestore.ShareableStore) database.SubRepoPermsStore {
				panic("unexpected invocation of MockSubRepoPermsStore.With")
			},
		},
	}
}

// NewMockSubRepoPermsStoreFrom creates a new mock of the
// MockSubRepoPermsStore interface. All methods delegate to the given
// implementation, unless overwritten.
func NewMockSubRepoPermsStoreFrom(i database.SubRepoPermsStore) *MockSubRepoPermsStore {
	return &MockSubRepoPermsStore{
		DoneFunc: &SubRepoPermsStoreDoneFunc{
			defaultHook: i.Done,
		},
		GetFunc: &SubRepoPermsStoreGetFunc{
			defaultHook: i.Get,
		},
		GetByUserFunc: &SubRepoPermsStoreGetByUserFunc{
			defaultHook: i.GetByUser,
		},
		TransactFunc: &SubRepoPermsStoreTransactFunc{
			defaultHook: i.Transact,
		},
		UpsertFunc: &SubRepoPermsStoreUpsertFunc{
			defaultHook: i.Upsert,
		},
		UpsertWithSpecFunc: &SubRepoPermsStoreUpsertWithSpecFunc{
			defaultHook: i.UpsertWithSpec,
		},
		WithFunc: &SubRepoPermsStoreWithFunc{
			defaultHook: i.With,
		},
	}
}

// SubRepoPermsStoreDoneFunc describes the behavior when the Done method of
// the parent MockSubRepoPermsStore instance is invoked.
type SubRepoPermsStoreDoneFunc struct {
	defaultHook func(error) error
	hooks       []func(error) error
	history     []SubRepoPermsStoreDoneFuncCall
	mutex       sync.Mutex
}

// Done delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.appendCall(SubRepoPermsStoreDoneFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Done method of the
// parent MockSubRepoPermsStore instance is invoked and the hook queue is
// empty.
func (f *SubRepoPermsStoreDoneFunc) SetDefaultHook(hook func(error) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Done method of the parent MockSubRepoPermsStore instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *SubRepoPermsStoreDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreDoneFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(error) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *SubRepoPermsStoreDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreDoneFunc) appendCall(r0 SubRepoPermsStoreDoneFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreDoneFuncCall objects
// describing the invocations of this function.
func (f *SubRepoPermsStoreDoneFunc) History() []SubRepoPermsStoreDoneFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreDoneFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreDoneFuncCall is an object that describes an invocation
// of method Done on an instance of MockSubRepoPermsStore.
type SubRepoPermsStoreDoneFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 error
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreDoneFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreDoneFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// SubRepoPermsStoreGetFunc describes the behavior when the Get method of
// the parent MockSubRepoPermsStore instance is invoked.
type SubRepoPermsStoreGetFunc struct {
	defaultHook func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error)
	hooks       []func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error)
	history     []SubRepoPermsStoreGetFuncCall
	mutex       sync.Mutex
}

// Get delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) Get(v0 context.Context, v1 int32, v2 api.RepoID) (*authz.SubRepoPermissions, error) {
	r0, r1 := m.GetFunc.nextHook()(v0, v1, v2)
	m.GetFunc.appendCall(SubRepoPermsStoreGetFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Get method of the
// parent MockSubRepoPermsStore instance is invoked and the hook queue is
// empty.
func (f *SubRepoPermsStoreGetFunc) SetDefaultHook(hook func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Get method of the parent MockSubRepoPermsStore instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *SubRepoPermsStoreGetFunc) PushHook(hook func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreGetFunc) SetDefaultReturn(r0 *authz.SubRepoPermissions, r1 error) {
	f.SetDefaultHook(func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreGetFunc) PushReturn(r0 *authz.SubRepoPermissions, r1 error) {
	f.PushHook(func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error) {
		return r0, r1
	})
}

func (f *SubRepoPermsStoreGetFunc) nextHook() func(context.Context, int32, api.RepoID) (*authz.SubRepoPermissions, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreGetFunc) appendCall(r0 SubRepoPermsStoreGetFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreGetFuncCall objects
// describing the invocations of this function.
func (f *SubRepoPermsStoreGetFunc) History() []SubRepoPermsStoreGetFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreGetFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreGetFuncCall is an object that describes an invocation of
// method Get on an instance of MockSubRepoPermsStore.
type SubRepoPermsStoreGetFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int32
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 api.RepoID
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *authz.SubRepoPermissions
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreGetFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreGetFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// SubRepoPermsStoreGetByUserFunc describes the behavior when the GetByUser
// method of the parent MockSubRepoPermsStore instance is invoked.
type SubRepoPermsStoreGetByUserFunc struct {
	defaultHook func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error)
	hooks       []func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error)
	history     []SubRepoPermsStoreGetByUserFuncCall
	mutex       sync.Mutex
}

// GetByUser delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) GetByUser(v0 context.Context, v1 int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
	r0, r1 := m.GetByUserFunc.nextHook()(v0, v1)
	m.GetByUserFunc.appendCall(SubRepoPermsStoreGetByUserFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetByUser method of
// the parent MockSubRepoPermsStore instance is invoked and the hook queue
// is empty.
func (f *SubRepoPermsStoreGetByUserFunc) SetDefaultHook(hook func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetByUser method of the parent MockSubRepoPermsStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *SubRepoPermsStoreGetByUserFunc) PushHook(hook func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreGetByUserFunc) SetDefaultReturn(r0 map[api.RepoName]authz.SubRepoPermissions, r1 error) {
	f.SetDefaultHook(func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreGetByUserFunc) PushReturn(r0 map[api.RepoName]authz.SubRepoPermissions, r1 error) {
	f.PushHook(func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
		return r0, r1
	})
}

func (f *SubRepoPermsStoreGetByUserFunc) nextHook() func(context.Context, int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreGetByUserFunc) appendCall(r0 SubRepoPermsStoreGetByUserFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreGetByUserFuncCall objects
// describing the invocations of this function.
func (f *SubRepoPermsStoreGetByUserFunc) History() []SubRepoPermsStoreGetByUserFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreGetByUserFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreGetByUserFuncCall is an object that describes an
// invocation of method GetByUser on an instance of MockSubRepoPermsStore.
type SubRepoPermsStoreGetByUserFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int32
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[api.RepoName]authz.SubRepoPermissions
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreGetByUserFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreGetByUserFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// SubRepoPermsStoreTransactFunc describes the behavior when the Transact
// method of the parent MockSubRepoPermsStore instance is invoked.
type SubRepoPermsStoreTransactFunc struct {
	defaultHook func(context.Context) (database.SubRepoPermsStore, error)
	hooks       []func(context.Context) (database.SubRepoPermsStore, error)
	history     []SubRepoPermsStoreTransactFuncCall
	mutex       sync.Mutex
}

// Transact delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) Transact(v0 context.Context) (database.SubRepoPermsStore, error) {
	r0, r1 := m.TransactFunc.nextHook()(v0)
	m.TransactFunc.appendCall(SubRepoPermsStoreTransactFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Transact method of
// the parent MockSubRepoPermsStore instance is invoked and the hook queue
// is empty.
func (f *SubRepoPermsStoreTransactFunc) SetDefaultHook(hook func(context.Context) (database.SubRepoPermsStore, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Transact method of the parent MockSubRepoPermsStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *SubRepoPermsStoreTransactFunc) PushHook(hook func(context.Context) (database.SubRepoPermsStore, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreTransactFunc) SetDefaultReturn(r0 database.SubRepoPermsStore, r1 error) {
	f.SetDefaultHook(func(context.Context) (database.SubRepoPermsStore, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreTransactFunc) PushReturn(r0 database.SubRepoPermsStore, r1 error) {
	f.PushHook(func(context.Context) (database.SubRepoPermsStore, error) {
		return r0, r1
	})
}

func (f *SubRepoPermsStoreTransactFunc) nextHook() func(context.Context) (database.SubRepoPermsStore, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreTransactFunc) appendCall(r0 SubRepoPermsStoreTransactFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreTransactFuncCall objects
// describing the invocations of this function.
func (f *SubRepoPermsStoreTransactFunc) History() []SubRepoPermsStoreTransactFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreTransactFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreTransactFuncCall is an object that describes an
// invocation of method Transact on an instance of MockSubRepoPermsStore.
type SubRepoPermsStoreTransactFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 database.SubRepoPermsStore
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreTransactFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreTransactFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// SubRepoPermsStoreUpsertFunc describes the behavior when the Upsert method
// of the parent MockSubRepoPermsStore instance is invoked.
type SubRepoPermsStoreUpsertFunc struct {
	defaultHook func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error
	hooks       []func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error
	history     []SubRepoPermsStoreUpsertFuncCall
	mutex       sync.Mutex
}

// Upsert delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) Upsert(v0 context.Context, v1 int32, v2 api.RepoID, v3 authz.SubRepoPermissions) error {
	r0 := m.UpsertFunc.nextHook()(v0, v1, v2, v3)
	m.UpsertFunc.appendCall(SubRepoPermsStoreUpsertFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Upsert method of the
// parent MockSubRepoPermsStore instance is invoked and the hook queue is
// empty.
func (f *SubRepoPermsStoreUpsertFunc) SetDefaultHook(hook func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Upsert method of the parent MockSubRepoPermsStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *SubRepoPermsStoreUpsertFunc) PushHook(hook func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreUpsertFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreUpsertFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error {
		return r0
	})
}

func (f *SubRepoPermsStoreUpsertFunc) nextHook() func(context.Context, int32, api.RepoID, authz.SubRepoPermissions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreUpsertFunc) appendCall(r0 SubRepoPermsStoreUpsertFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreUpsertFuncCall objects
// describing the invocations of this function.
func (f *SubRepoPermsStoreUpsertFunc) History() []SubRepoPermsStoreUpsertFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreUpsertFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreUpsertFuncCall is an object that describes an invocation
// of method Upsert on an instance of MockSubRepoPermsStore.
type SubRepoPermsStoreUpsertFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int32
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 api.RepoID
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 authz.SubRepoPermissions
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreUpsertFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreUpsertFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// SubRepoPermsStoreUpsertWithSpecFunc describes the behavior when the
// UpsertWithSpec method of the parent MockSubRepoPermsStore instance is
// invoked.
type SubRepoPermsStoreUpsertWithSpecFunc struct {
	defaultHook func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error
	hooks       []func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error
	history     []SubRepoPermsStoreUpsertWithSpecFuncCall
	mutex       sync.Mutex
}

// UpsertWithSpec delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) UpsertWithSpec(v0 context.Context, v1 int32, v2 api.ExternalRepoSpec, v3 authz.SubRepoPermissions) error {
	r0 := m.UpsertWithSpecFunc.nextHook()(v0, v1, v2, v3)
	m.UpsertWithSpecFunc.appendCall(SubRepoPermsStoreUpsertWithSpecFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the UpsertWithSpec
// method of the parent MockSubRepoPermsStore instance is invoked and the
// hook queue is empty.
func (f *SubRepoPermsStoreUpsertWithSpecFunc) SetDefaultHook(hook func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// UpsertWithSpec method of the parent MockSubRepoPermsStore instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *SubRepoPermsStoreUpsertWithSpecFunc) PushHook(hook func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreUpsertWithSpecFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreUpsertWithSpecFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error {
		return r0
	})
}

func (f *SubRepoPermsStoreUpsertWithSpecFunc) nextHook() func(context.Context, int32, api.ExternalRepoSpec, authz.SubRepoPermissions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreUpsertWithSpecFunc) appendCall(r0 SubRepoPermsStoreUpsertWithSpecFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreUpsertWithSpecFuncCall
// objects describing the invocations of this function.
func (f *SubRepoPermsStoreUpsertWithSpecFunc) History() []SubRepoPermsStoreUpsertWithSpecFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreUpsertWithSpecFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreUpsertWithSpecFuncCall is an object that describes an
// invocation of method UpsertWithSpec on an instance of
// MockSubRepoPermsStore.
type SubRepoPermsStoreUpsertWithSpecFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int32
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 api.ExternalRepoSpec
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 authz.SubRepoPermissions
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreUpsertWithSpecFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreUpsertWithSpecFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// SubRepoPermsStoreWithFunc describes the behavior when the With method of
// the parent MockSubRepoPermsStore instance is invoked.
type SubRepoPermsStoreWithFunc struct {
	defaultHook func(basestore.ShareableStore) database.SubRepoPermsStore
	hooks       []func(basestore.ShareableStore) database.SubRepoPermsStore
	history     []SubRepoPermsStoreWithFuncCall
	mutex       sync.Mutex
}

// With delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockSubRepoPermsStore) With(v0 basestore.ShareableStore) database.SubRepoPermsStore {
	r0 := m.WithFunc.nextHook()(v0)
	m.WithFunc.appendCall(SubRepoPermsStoreWithFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the With method of the
// parent MockSubRepoPermsStore instance is invoked and the hook queue is
// empty.
func (f *SubRepoPermsStoreWithFunc) SetDefaultHook(hook func(basestore.ShareableStore) database.SubRepoPermsStore) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// With method of the parent MockSubRepoPermsStore instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *SubRepoPermsStoreWithFunc) PushHook(hook func(basestore.ShareableStore) database.SubRepoPermsStore) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *SubRepoPermsStoreWithFunc) SetDefaultReturn(r0 database.SubRepoPermsStore) {
	f.SetDefaultHook(func(basestore.ShareableStore) database.SubRepoPermsStore {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *SubRepoPermsStoreWithFunc) PushReturn(r0 database.SubRepoPermsStore) {
	f.PushHook(func(basestore.ShareableStore) database.SubRepoPermsStore {
		return r0
	})
}

func (f *SubRepoPermsStoreWithFunc) nextHook() func(basestore.ShareableStore) database.SubRepoPermsStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SubRepoPermsStoreWithFunc) appendCall(r0 SubRepoPermsStoreWithFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of SubRepoPermsStoreWithFuncCall objects
// describing the invocations of this function.
func (f *SubRepoPermsStoreWithFunc) History() []SubRepoPermsStoreWithFuncCall {
	f.mutex.Lock()
	history := make([]SubRepoPermsStoreWithFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SubRepoPermsStoreWithFuncCall is an object that describes an invocation
// of method With on an instance of MockSubRepoPermsStore.
type SubRepoPermsStoreWithFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 basestore.ShareableStore
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 database.SubRepoPermsStore
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c SubRepoPermsStoreWithFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c SubRepoPermsStoreWithFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
