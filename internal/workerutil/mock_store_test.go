// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package workerutil

import (
	"context"
	"sync"
)

// MockStore is a mock implementation of the Store interface (from the
// package github.com/sourcegraph/sourcegraph/internal/workerutil) used for
// unit testing.
type MockStore struct {
	// DequeueFunc is an instance of a mock function object controlling the
	// behavior of the method Dequeue.
	DequeueFunc *StoreDequeueFunc
	// DoneFunc is an instance of a mock function object controlling the
	// behavior of the method Done.
	DoneFunc *StoreDoneFunc
	// MarkCompleteFunc is an instance of a mock function object controlling
	// the behavior of the method MarkComplete.
	MarkCompleteFunc *StoreMarkCompleteFunc
	// MarkErroredFunc is an instance of a mock function object controlling
	// the behavior of the method MarkErrored.
	MarkErroredFunc *StoreMarkErroredFunc
	// QueuedCountFunc is an instance of a mock function object controlling
	// the behavior of the method QueuedCount.
	QueuedCountFunc *StoreQueuedCountFunc
	// SetLogContentsFunc is an instance of a mock function object
	// controlling the behavior of the method SetLogContents.
	SetLogContentsFunc *StoreSetLogContentsFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		DequeueFunc: &StoreDequeueFunc{
			defaultHook: func(context.Context, interface{}) (Record, Store, bool, error) {
				return nil, nil, false, nil
			},
		},
		DoneFunc: &StoreDoneFunc{
			defaultHook: func(error) error {
				return nil
			},
		},
		MarkCompleteFunc: &StoreMarkCompleteFunc{
			defaultHook: func(context.Context, int) (bool, error) {
				return false, nil
			},
		},
		MarkErroredFunc: &StoreMarkErroredFunc{
			defaultHook: func(context.Context, int, string) (bool, error) {
				return false, nil
			},
		},
		QueuedCountFunc: &StoreQueuedCountFunc{
			defaultHook: func(context.Context, interface{}) (int, error) {
				return 0, nil
			},
		},
		SetLogContentsFunc: &StoreSetLogContentsFunc{
			defaultHook: func(context.Context, int, string) error {
				return nil
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		DequeueFunc: &StoreDequeueFunc{
			defaultHook: i.Dequeue,
		},
		DoneFunc: &StoreDoneFunc{
			defaultHook: i.Done,
		},
		MarkCompleteFunc: &StoreMarkCompleteFunc{
			defaultHook: i.MarkComplete,
		},
		MarkErroredFunc: &StoreMarkErroredFunc{
			defaultHook: i.MarkErrored,
		},
		QueuedCountFunc: &StoreQueuedCountFunc{
			defaultHook: i.QueuedCount,
		},
		SetLogContentsFunc: &StoreSetLogContentsFunc{
			defaultHook: i.SetLogContents,
		},
	}
}

// StoreDequeueFunc describes the behavior when the Dequeue method of the
// parent MockStore instance is invoked.
type StoreDequeueFunc struct {
	defaultHook func(context.Context, interface{}) (Record, Store, bool, error)
	hooks       []func(context.Context, interface{}) (Record, Store, bool, error)
	history     []StoreDequeueFuncCall
	mutex       sync.Mutex
}

// Dequeue delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Dequeue(v0 context.Context, v1 interface{}) (Record, Store, bool, error) {
	r0, r1, r2, r3 := m.DequeueFunc.nextHook()(v0, v1)
	m.DequeueFunc.appendCall(StoreDequeueFuncCall{v0, v1, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefaultHook sets function that is called when the Dequeue method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreDequeueFunc) SetDefaultHook(hook func(context.Context, interface{}) (Record, Store, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Dequeue method of the parent MockStore instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreDequeueFunc) PushHook(hook func(context.Context, interface{}) (Record, Store, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreDequeueFunc) SetDefaultReturn(r0 Record, r1 Store, r2 bool, r3 error) {
	f.SetDefaultHook(func(context.Context, interface{}) (Record, Store, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreDequeueFunc) PushReturn(r0 Record, r1 Store, r2 bool, r3 error) {
	f.PushHook(func(context.Context, interface{}) (Record, Store, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *StoreDequeueFunc) nextHook() func(context.Context, interface{}) (Record, Store, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDequeueFunc) appendCall(r0 StoreDequeueFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDequeueFuncCall objects describing the
// invocations of this function.
func (f *StoreDequeueFunc) History() []StoreDequeueFuncCall {
	f.mutex.Lock()
	history := make([]StoreDequeueFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDequeueFuncCall is an object that describes an invocation of method
// Dequeue on an instance of MockStore.
type StoreDequeueFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 interface{}
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 Record
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 Store
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 bool
	// Result3 is the value of the 4th result returned from this method
	// invocation.
	Result3 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDequeueFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDequeueFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// StoreDoneFunc describes the behavior when the Done method of the parent
// MockStore instance is invoked.
type StoreDoneFunc struct {
	defaultHook func(error) error
	hooks       []func(error) error
	history     []StoreDoneFuncCall
	mutex       sync.Mutex
}

// Done delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.appendCall(StoreDoneFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Done method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreDoneFunc) SetDefaultHook(hook func(error) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Done method of the parent MockStore instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreDoneFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(error) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *StoreDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDoneFunc) appendCall(r0 StoreDoneFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDoneFuncCall objects describing the
// invocations of this function.
func (f *StoreDoneFunc) History() []StoreDoneFuncCall {
	f.mutex.Lock()
	history := make([]StoreDoneFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDoneFuncCall is an object that describes an invocation of method
// Done on an instance of MockStore.
type StoreDoneFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 error
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDoneFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDoneFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreMarkCompleteFunc describes the behavior when the MarkComplete method
// of the parent MockStore instance is invoked.
type StoreMarkCompleteFunc struct {
	defaultHook func(context.Context, int) (bool, error)
	hooks       []func(context.Context, int) (bool, error)
	history     []StoreMarkCompleteFuncCall
	mutex       sync.Mutex
}

// MarkComplete delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) MarkComplete(v0 context.Context, v1 int) (bool, error) {
	r0, r1 := m.MarkCompleteFunc.nextHook()(v0, v1)
	m.MarkCompleteFunc.appendCall(StoreMarkCompleteFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the MarkComplete method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreMarkCompleteFunc) SetDefaultHook(hook func(context.Context, int) (bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// MarkComplete method of the parent MockStore instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreMarkCompleteFunc) PushHook(hook func(context.Context, int) (bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreMarkCompleteFunc) SetDefaultReturn(r0 bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreMarkCompleteFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMarkCompleteFunc) nextHook() func(context.Context, int) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMarkCompleteFunc) appendCall(r0 StoreMarkCompleteFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreMarkCompleteFuncCall objects
// describing the invocations of this function.
func (f *StoreMarkCompleteFunc) History() []StoreMarkCompleteFuncCall {
	f.mutex.Lock()
	history := make([]StoreMarkCompleteFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMarkCompleteFuncCall is an object that describes an invocation of
// method MarkComplete on an instance of MockStore.
type StoreMarkCompleteFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreMarkCompleteFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreMarkCompleteFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreMarkErroredFunc describes the behavior when the MarkErrored method
// of the parent MockStore instance is invoked.
type StoreMarkErroredFunc struct {
	defaultHook func(context.Context, int, string) (bool, error)
	hooks       []func(context.Context, int, string) (bool, error)
	history     []StoreMarkErroredFuncCall
	mutex       sync.Mutex
}

// MarkErrored delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) MarkErrored(v0 context.Context, v1 int, v2 string) (bool, error) {
	r0, r1 := m.MarkErroredFunc.nextHook()(v0, v1, v2)
	m.MarkErroredFunc.appendCall(StoreMarkErroredFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the MarkErrored method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreMarkErroredFunc) SetDefaultHook(hook func(context.Context, int, string) (bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// MarkErrored method of the parent MockStore instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreMarkErroredFunc) PushHook(hook func(context.Context, int, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreMarkErroredFunc) SetDefaultReturn(r0 bool, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreMarkErroredFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

func (f *StoreMarkErroredFunc) nextHook() func(context.Context, int, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreMarkErroredFunc) appendCall(r0 StoreMarkErroredFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreMarkErroredFuncCall objects describing
// the invocations of this function.
func (f *StoreMarkErroredFunc) History() []StoreMarkErroredFuncCall {
	f.mutex.Lock()
	history := make([]StoreMarkErroredFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreMarkErroredFuncCall is an object that describes an invocation of
// method MarkErrored on an instance of MockStore.
type StoreMarkErroredFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreMarkErroredFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreMarkErroredFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreQueuedCountFunc describes the behavior when the QueuedCount method
// of the parent MockStore instance is invoked.
type StoreQueuedCountFunc struct {
	defaultHook func(context.Context, interface{}) (int, error)
	hooks       []func(context.Context, interface{}) (int, error)
	history     []StoreQueuedCountFuncCall
	mutex       sync.Mutex
}

// QueuedCount delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) QueuedCount(v0 context.Context, v1 interface{}) (int, error) {
	r0, r1 := m.QueuedCountFunc.nextHook()(v0, v1)
	m.QueuedCountFunc.appendCall(StoreQueuedCountFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the QueuedCount method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreQueuedCountFunc) SetDefaultHook(hook func(context.Context, interface{}) (int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// QueuedCount method of the parent MockStore instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreQueuedCountFunc) PushHook(hook func(context.Context, interface{}) (int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreQueuedCountFunc) SetDefaultReturn(r0 int, r1 error) {
	f.SetDefaultHook(func(context.Context, interface{}) (int, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreQueuedCountFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, interface{}) (int, error) {
		return r0, r1
	})
}

func (f *StoreQueuedCountFunc) nextHook() func(context.Context, interface{}) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreQueuedCountFunc) appendCall(r0 StoreQueuedCountFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreQueuedCountFuncCall objects describing
// the invocations of this function.
func (f *StoreQueuedCountFunc) History() []StoreQueuedCountFuncCall {
	f.mutex.Lock()
	history := make([]StoreQueuedCountFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreQueuedCountFuncCall is an object that describes an invocation of
// method QueuedCount on an instance of MockStore.
type StoreQueuedCountFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 interface{}
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreQueuedCountFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreQueuedCountFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreSetLogContentsFunc describes the behavior when the SetLogContents
// method of the parent MockStore instance is invoked.
type StoreSetLogContentsFunc struct {
	defaultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []StoreSetLogContentsFuncCall
	mutex       sync.Mutex
}

// SetLogContents delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) SetLogContents(v0 context.Context, v1 int, v2 string) error {
	r0 := m.SetLogContentsFunc.nextHook()(v0, v1, v2)
	m.SetLogContentsFunc.appendCall(StoreSetLogContentsFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the SetLogContents
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreSetLogContentsFunc) SetDefaultHook(hook func(context.Context, int, string) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SetLogContents method of the parent MockStore instance inovkes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreSetLogContentsFunc) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreSetLogContentsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreSetLogContentsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *StoreSetLogContentsFunc) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSetLogContentsFunc) appendCall(r0 StoreSetLogContentsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreSetLogContentsFuncCall objects
// describing the invocations of this function.
func (f *StoreSetLogContentsFunc) History() []StoreSetLogContentsFuncCall {
	f.mutex.Lock()
	history := make([]StoreSetLogContentsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSetLogContentsFuncCall is an object that describes an invocation of
// method SetLogContents on an instance of MockStore.
type StoreSetLogContentsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreSetLogContentsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreSetLogContentsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
