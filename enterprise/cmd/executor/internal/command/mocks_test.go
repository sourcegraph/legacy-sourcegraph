// Code generated by go-mockgen 1.3.2; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package command

import (
	"context"
	"sync"

	workerutil "github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// MockExecutionLogEntryStore is a mock implementation of the
// ExecutionLogEntryStore interface (from the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command)
// used for unit testing.
type MockExecutionLogEntryStore struct {
	// AddExecutionLogEntryFunc is an instance of a mock function object
	// controlling the behavior of the method AddExecutionLogEntry.
	AddExecutionLogEntryFunc *ExecutionLogEntryStoreAddExecutionLogEntryFunc
	// UpdateExecutionLogEntryFunc is an instance of a mock function object
	// controlling the behavior of the method UpdateExecutionLogEntry.
	UpdateExecutionLogEntryFunc *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc
}

// NewMockExecutionLogEntryStore creates a new mock of the
// ExecutionLogEntryStore interface. All methods return zero values for all
// results, unless overwritten.
func NewMockExecutionLogEntryStore() *MockExecutionLogEntryStore {
	return &MockExecutionLogEntryStore{
		AddExecutionLogEntryFunc: &ExecutionLogEntryStoreAddExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, workerutil.ExecutionLogEntry) (r0 int, r1 error) {
				return
			},
		},
		UpdateExecutionLogEntryFunc: &ExecutionLogEntryStoreUpdateExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, int, workerutil.ExecutionLogEntry) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockExecutionLogEntryStore creates a new mock of the
// ExecutionLogEntryStore interface. All methods panic on invocation, unless
// overwritten.
func NewStrictMockExecutionLogEntryStore() *MockExecutionLogEntryStore {
	return &MockExecutionLogEntryStore{
		AddExecutionLogEntryFunc: &ExecutionLogEntryStoreAddExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, workerutil.ExecutionLogEntry) (int, error) {
				panic("unexpected invocation of MockExecutionLogEntryStore.AddExecutionLogEntry")
			},
		},
		UpdateExecutionLogEntryFunc: &ExecutionLogEntryStoreUpdateExecutionLogEntryFunc{
			defaultHook: func(context.Context, int, int, workerutil.ExecutionLogEntry) error {
				panic("unexpected invocation of MockExecutionLogEntryStore.UpdateExecutionLogEntry")
			},
		},
	}
}

// NewMockExecutionLogEntryStoreFrom creates a new mock of the
// MockExecutionLogEntryStore interface. All methods delegate to the given
// implementation, unless overwritten.
func NewMockExecutionLogEntryStoreFrom(i ExecutionLogEntryStore) *MockExecutionLogEntryStore {
	return &MockExecutionLogEntryStore{
		AddExecutionLogEntryFunc: &ExecutionLogEntryStoreAddExecutionLogEntryFunc{
			defaultHook: i.AddExecutionLogEntry,
		},
		UpdateExecutionLogEntryFunc: &ExecutionLogEntryStoreUpdateExecutionLogEntryFunc{
			defaultHook: i.UpdateExecutionLogEntry,
		},
	}
}

// ExecutionLogEntryStoreAddExecutionLogEntryFunc describes the behavior
// when the AddExecutionLogEntry method of the parent
// MockExecutionLogEntryStore instance is invoked.
type ExecutionLogEntryStoreAddExecutionLogEntryFunc struct {
	defaultHook func(context.Context, int, workerutil.ExecutionLogEntry) (int, error)
	hooks       []func(context.Context, int, workerutil.ExecutionLogEntry) (int, error)
	history     []ExecutionLogEntryStoreAddExecutionLogEntryFuncCall
	mutex       sync.Mutex
}

// AddExecutionLogEntry delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockExecutionLogEntryStore) AddExecutionLogEntry(v0 context.Context, v1 int, v2 workerutil.ExecutionLogEntry) (int, error) {
	r0, r1 := m.AddExecutionLogEntryFunc.nextHook()(v0, v1, v2)
	m.AddExecutionLogEntryFunc.appendCall(ExecutionLogEntryStoreAddExecutionLogEntryFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the AddExecutionLogEntry
// method of the parent MockExecutionLogEntryStore instance is invoked and
// the hook queue is empty.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) SetDefaultHook(hook func(context.Context, int, workerutil.ExecutionLogEntry) (int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// AddExecutionLogEntry method of the parent MockExecutionLogEntryStore
// instance invokes the hook at the front of the queue and discards it.
// After the queue is empty, the default hook function is invoked for any
// future action.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) PushHook(hook func(context.Context, int, workerutil.ExecutionLogEntry) (int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) SetDefaultReturn(r0 int, r1 error) {
	f.SetDefaultHook(func(context.Context, int, workerutil.ExecutionLogEntry) (int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, int, workerutil.ExecutionLogEntry) (int, error) {
		return r0, r1
	})
}

func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) nextHook() func(context.Context, int, workerutil.ExecutionLogEntry) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) appendCall(r0 ExecutionLogEntryStoreAddExecutionLogEntryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// ExecutionLogEntryStoreAddExecutionLogEntryFuncCall objects describing the
// invocations of this function.
func (f *ExecutionLogEntryStoreAddExecutionLogEntryFunc) History() []ExecutionLogEntryStoreAddExecutionLogEntryFuncCall {
	f.mutex.Lock()
	history := make([]ExecutionLogEntryStoreAddExecutionLogEntryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExecutionLogEntryStoreAddExecutionLogEntryFuncCall is an object that
// describes an invocation of method AddExecutionLogEntry on an instance of
// MockExecutionLogEntryStore.
type ExecutionLogEntryStoreAddExecutionLogEntryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 workerutil.ExecutionLogEntry
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ExecutionLogEntryStoreAddExecutionLogEntryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ExecutionLogEntryStoreAddExecutionLogEntryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// ExecutionLogEntryStoreUpdateExecutionLogEntryFunc describes the behavior
// when the UpdateExecutionLogEntry method of the parent
// MockExecutionLogEntryStore instance is invoked.
type ExecutionLogEntryStoreUpdateExecutionLogEntryFunc struct {
	defaultHook func(context.Context, int, int, workerutil.ExecutionLogEntry) error
	hooks       []func(context.Context, int, int, workerutil.ExecutionLogEntry) error
	history     []ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall
	mutex       sync.Mutex
}

// UpdateExecutionLogEntry delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockExecutionLogEntryStore) UpdateExecutionLogEntry(v0 context.Context, v1 int, v2 int, v3 workerutil.ExecutionLogEntry) error {
	r0 := m.UpdateExecutionLogEntryFunc.nextHook()(v0, v1, v2, v3)
	m.UpdateExecutionLogEntryFunc.appendCall(ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the
// UpdateExecutionLogEntry method of the parent MockExecutionLogEntryStore
// instance is invoked and the hook queue is empty.
func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) SetDefaultHook(hook func(context.Context, int, int, workerutil.ExecutionLogEntry) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// UpdateExecutionLogEntry method of the parent MockExecutionLogEntryStore
// instance invokes the hook at the front of the queue and discards it.
// After the queue is empty, the default hook function is invoked for any
// future action.
func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) PushHook(hook func(context.Context, int, int, workerutil.ExecutionLogEntry) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, int, int, workerutil.ExecutionLogEntry) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, int, workerutil.ExecutionLogEntry) error {
		return r0
	})
}

func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) nextHook() func(context.Context, int, int, workerutil.ExecutionLogEntry) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) appendCall(r0 ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall objects describing
// the invocations of this function.
func (f *ExecutionLogEntryStoreUpdateExecutionLogEntryFunc) History() []ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall {
	f.mutex.Lock()
	history := make([]ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall is an object that
// describes an invocation of method UpdateExecutionLogEntry on an instance
// of MockExecutionLogEntryStore.
type ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 workerutil.ExecutionLogEntry
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ExecutionLogEntryStoreUpdateExecutionLogEntryFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// MockCommandRunner is a mock implementation of the commandRunner interface
// (from the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command)
// used for unit testing.
type MockCommandRunner struct {
	// RunCommandFunc is an instance of a mock function object controlling
	// the behavior of the method RunCommand.
	RunCommandFunc *CommandRunnerRunCommandFunc
}

// NewMockCommandRunner creates a new mock of the commandRunner interface.
// All methods return zero values for all results, unless overwritten.
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		RunCommandFunc: &CommandRunnerRunCommandFunc{
			defaultHook: func(context.Context, command, *Logger) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockCommandRunner creates a new mock of the commandRunner
// interface. All methods panic on invocation, unless overwritten.
func NewStrictMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		RunCommandFunc: &CommandRunnerRunCommandFunc{
			defaultHook: func(context.Context, command, *Logger) error {
				panic("unexpected invocation of MockCommandRunner.RunCommand")
			},
		},
	}
}

// surrogateMockCommandRunner is a copy of the commandRunner interface (from
// the package
// github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command).
// It is redefined here as it is unexported in the source package.
type surrogateMockCommandRunner interface {
	RunCommand(context.Context, command, *Logger) error
}

// NewMockCommandRunnerFrom creates a new mock of the MockCommandRunner
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockCommandRunnerFrom(i surrogateMockCommandRunner) *MockCommandRunner {
	return &MockCommandRunner{
		RunCommandFunc: &CommandRunnerRunCommandFunc{
			defaultHook: i.RunCommand,
		},
	}
}

// CommandRunnerRunCommandFunc describes the behavior when the RunCommand
// method of the parent MockCommandRunner instance is invoked.
type CommandRunnerRunCommandFunc struct {
	defaultHook func(context.Context, command, *Logger) error
	hooks       []func(context.Context, command, *Logger) error
	history     []CommandRunnerRunCommandFuncCall
	mutex       sync.Mutex
}

// RunCommand delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCommandRunner) RunCommand(v0 context.Context, v1 command, v2 *Logger) error {
	r0 := m.RunCommandFunc.nextHook()(v0, v1, v2)
	m.RunCommandFunc.appendCall(CommandRunnerRunCommandFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the RunCommand method of
// the parent MockCommandRunner instance is invoked and the hook queue is
// empty.
func (f *CommandRunnerRunCommandFunc) SetDefaultHook(hook func(context.Context, command, *Logger) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RunCommand method of the parent MockCommandRunner instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CommandRunnerRunCommandFunc) PushHook(hook func(context.Context, command, *Logger) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *CommandRunnerRunCommandFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, command, *Logger) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *CommandRunnerRunCommandFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, command, *Logger) error {
		return r0
	})
}

func (f *CommandRunnerRunCommandFunc) nextHook() func(context.Context, command, *Logger) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CommandRunnerRunCommandFunc) appendCall(r0 CommandRunnerRunCommandFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CommandRunnerRunCommandFuncCall objects
// describing the invocations of this function.
func (f *CommandRunnerRunCommandFunc) History() []CommandRunnerRunCommandFuncCall {
	f.mutex.Lock()
	history := make([]CommandRunnerRunCommandFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CommandRunnerRunCommandFuncCall is an object that describes an invocation
// of method RunCommand on an instance of MockCommandRunner.
type CommandRunnerRunCommandFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 command
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 *Logger
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CommandRunnerRunCommandFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CommandRunnerRunCommandFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
