// Code generated by go-mockgen 1.3.4; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package team

import (
	"context"
	"sync"
)

// MockTeammateResolver is a mock implementation of the TeammateResolver
// interface (from the package github.com/sourcegraph/sourcegraph/dev/team)
// used for unit testing.
type MockTeammateResolver struct {
	// ResolveByCommitAuthorFunc is an instance of a mock function object
	// controlling the behavior of the method ResolveByCommitAuthor.
	ResolveByCommitAuthorFunc *TeammateResolverResolveByCommitAuthorFunc
	// ResolveByGitHubHandleFunc is an instance of a mock function object
	// controlling the behavior of the method ResolveByGitHubHandle.
	ResolveByGitHubHandleFunc *TeammateResolverResolveByGitHubHandleFunc
	// ResolveByNameFunc is an instance of a mock function object
	// controlling the behavior of the method ResolveByName.
	ResolveByNameFunc *TeammateResolverResolveByNameFunc
}

// NewMockTeammateResolver creates a new mock of the TeammateResolver
// interface. All methods return zero values for all results, unless
// overwritten.
func NewMockTeammateResolver() *MockTeammateResolver {
	return &MockTeammateResolver{
		ResolveByCommitAuthorFunc: &TeammateResolverResolveByCommitAuthorFunc{
			defaultHook: func(context.Context, string, string, string) (r0 *Teammate, r1 error) {
				return
			},
		},
		ResolveByGitHubHandleFunc: &TeammateResolverResolveByGitHubHandleFunc{
			defaultHook: func(context.Context, string) (r0 *Teammate, r1 error) {
				return
			},
		},
		ResolveByNameFunc: &TeammateResolverResolveByNameFunc{
			defaultHook: func(context.Context, string) (r0 *Teammate, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockTeammateResolver creates a new mock of the TeammateResolver
// interface. All methods panic on invocation, unless overwritten.
func NewStrictMockTeammateResolver() *MockTeammateResolver {
	return &MockTeammateResolver{
		ResolveByCommitAuthorFunc: &TeammateResolverResolveByCommitAuthorFunc{
			defaultHook: func(context.Context, string, string, string) (*Teammate, error) {
				panic("unexpected invocation of MockTeammateResolver.ResolveByCommitAuthor")
			},
		},
		ResolveByGitHubHandleFunc: &TeammateResolverResolveByGitHubHandleFunc{
			defaultHook: func(context.Context, string) (*Teammate, error) {
				panic("unexpected invocation of MockTeammateResolver.ResolveByGitHubHandle")
			},
		},
		ResolveByNameFunc: &TeammateResolverResolveByNameFunc{
			defaultHook: func(context.Context, string) (*Teammate, error) {
				panic("unexpected invocation of MockTeammateResolver.ResolveByName")
			},
		},
	}
}

// NewMockTeammateResolverFrom creates a new mock of the
// MockTeammateResolver interface. All methods delegate to the given
// implementation, unless overwritten.
func NewMockTeammateResolverFrom(i TeammateResolver) *MockTeammateResolver {
	return &MockTeammateResolver{
		ResolveByCommitAuthorFunc: &TeammateResolverResolveByCommitAuthorFunc{
			defaultHook: i.ResolveByCommitAuthor,
		},
		ResolveByGitHubHandleFunc: &TeammateResolverResolveByGitHubHandleFunc{
			defaultHook: i.ResolveByGitHubHandle,
		},
		ResolveByNameFunc: &TeammateResolverResolveByNameFunc{
			defaultHook: i.ResolveByName,
		},
	}
}

// TeammateResolverResolveByCommitAuthorFunc describes the behavior when the
// ResolveByCommitAuthor method of the parent MockTeammateResolver instance
// is invoked.
type TeammateResolverResolveByCommitAuthorFunc struct {
	defaultHook func(context.Context, string, string, string) (*Teammate, error)
	hooks       []func(context.Context, string, string, string) (*Teammate, error)
	history     []TeammateResolverResolveByCommitAuthorFuncCall
	mutex       sync.Mutex
}

// ResolveByCommitAuthor delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockTeammateResolver) ResolveByCommitAuthor(v0 context.Context, v1 string, v2 string, v3 string) (*Teammate, error) {
	r0, r1 := m.ResolveByCommitAuthorFunc.nextHook()(v0, v1, v2, v3)
	m.ResolveByCommitAuthorFunc.appendCall(TeammateResolverResolveByCommitAuthorFuncCall{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// ResolveByCommitAuthor method of the parent MockTeammateResolver instance
// is invoked and the hook queue is empty.
func (f *TeammateResolverResolveByCommitAuthorFunc) SetDefaultHook(hook func(context.Context, string, string, string) (*Teammate, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// ResolveByCommitAuthor method of the parent MockTeammateResolver instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *TeammateResolverResolveByCommitAuthorFunc) PushHook(hook func(context.Context, string, string, string) (*Teammate, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *TeammateResolverResolveByCommitAuthorFunc) SetDefaultReturn(r0 *Teammate, r1 error) {
	f.SetDefaultHook(func(context.Context, string, string, string) (*Teammate, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *TeammateResolverResolveByCommitAuthorFunc) PushReturn(r0 *Teammate, r1 error) {
	f.PushHook(func(context.Context, string, string, string) (*Teammate, error) {
		return r0, r1
	})
}

func (f *TeammateResolverResolveByCommitAuthorFunc) nextHook() func(context.Context, string, string, string) (*Teammate, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *TeammateResolverResolveByCommitAuthorFunc) appendCall(r0 TeammateResolverResolveByCommitAuthorFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// TeammateResolverResolveByCommitAuthorFuncCall objects describing the
// invocations of this function.
func (f *TeammateResolverResolveByCommitAuthorFunc) History() []TeammateResolverResolveByCommitAuthorFuncCall {
	f.mutex.Lock()
	history := make([]TeammateResolverResolveByCommitAuthorFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// TeammateResolverResolveByCommitAuthorFuncCall is an object that describes
// an invocation of method ResolveByCommitAuthor on an instance of
// MockTeammateResolver.
type TeammateResolverResolveByCommitAuthorFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *Teammate
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c TeammateResolverResolveByCommitAuthorFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c TeammateResolverResolveByCommitAuthorFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// TeammateResolverResolveByGitHubHandleFunc describes the behavior when the
// ResolveByGitHubHandle method of the parent MockTeammateResolver instance
// is invoked.
type TeammateResolverResolveByGitHubHandleFunc struct {
	defaultHook func(context.Context, string) (*Teammate, error)
	hooks       []func(context.Context, string) (*Teammate, error)
	history     []TeammateResolverResolveByGitHubHandleFuncCall
	mutex       sync.Mutex
}

// ResolveByGitHubHandle delegates to the next hook function in the queue
// and stores the parameter and result values of this invocation.
func (m *MockTeammateResolver) ResolveByGitHubHandle(v0 context.Context, v1 string) (*Teammate, error) {
	r0, r1 := m.ResolveByGitHubHandleFunc.nextHook()(v0, v1)
	m.ResolveByGitHubHandleFunc.appendCall(TeammateResolverResolveByGitHubHandleFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the
// ResolveByGitHubHandle method of the parent MockTeammateResolver instance
// is invoked and the hook queue is empty.
func (f *TeammateResolverResolveByGitHubHandleFunc) SetDefaultHook(hook func(context.Context, string) (*Teammate, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// ResolveByGitHubHandle method of the parent MockTeammateResolver instance
// invokes the hook at the front of the queue and discards it. After the
// queue is empty, the default hook function is invoked for any future
// action.
func (f *TeammateResolverResolveByGitHubHandleFunc) PushHook(hook func(context.Context, string) (*Teammate, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *TeammateResolverResolveByGitHubHandleFunc) SetDefaultReturn(r0 *Teammate, r1 error) {
	f.SetDefaultHook(func(context.Context, string) (*Teammate, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *TeammateResolverResolveByGitHubHandleFunc) PushReturn(r0 *Teammate, r1 error) {
	f.PushHook(func(context.Context, string) (*Teammate, error) {
		return r0, r1
	})
}

func (f *TeammateResolverResolveByGitHubHandleFunc) nextHook() func(context.Context, string) (*Teammate, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *TeammateResolverResolveByGitHubHandleFunc) appendCall(r0 TeammateResolverResolveByGitHubHandleFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of
// TeammateResolverResolveByGitHubHandleFuncCall objects describing the
// invocations of this function.
func (f *TeammateResolverResolveByGitHubHandleFunc) History() []TeammateResolverResolveByGitHubHandleFuncCall {
	f.mutex.Lock()
	history := make([]TeammateResolverResolveByGitHubHandleFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// TeammateResolverResolveByGitHubHandleFuncCall is an object that describes
// an invocation of method ResolveByGitHubHandle on an instance of
// MockTeammateResolver.
type TeammateResolverResolveByGitHubHandleFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *Teammate
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c TeammateResolverResolveByGitHubHandleFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c TeammateResolverResolveByGitHubHandleFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// TeammateResolverResolveByNameFunc describes the behavior when the
// ResolveByName method of the parent MockTeammateResolver instance is
// invoked.
type TeammateResolverResolveByNameFunc struct {
	defaultHook func(context.Context, string) (*Teammate, error)
	hooks       []func(context.Context, string) (*Teammate, error)
	history     []TeammateResolverResolveByNameFuncCall
	mutex       sync.Mutex
}

// ResolveByName delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockTeammateResolver) ResolveByName(v0 context.Context, v1 string) (*Teammate, error) {
	r0, r1 := m.ResolveByNameFunc.nextHook()(v0, v1)
	m.ResolveByNameFunc.appendCall(TeammateResolverResolveByNameFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the ResolveByName method
// of the parent MockTeammateResolver instance is invoked and the hook queue
// is empty.
func (f *TeammateResolverResolveByNameFunc) SetDefaultHook(hook func(context.Context, string) (*Teammate, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// ResolveByName method of the parent MockTeammateResolver instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *TeammateResolverResolveByNameFunc) PushHook(hook func(context.Context, string) (*Teammate, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *TeammateResolverResolveByNameFunc) SetDefaultReturn(r0 *Teammate, r1 error) {
	f.SetDefaultHook(func(context.Context, string) (*Teammate, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *TeammateResolverResolveByNameFunc) PushReturn(r0 *Teammate, r1 error) {
	f.PushHook(func(context.Context, string) (*Teammate, error) {
		return r0, r1
	})
}

func (f *TeammateResolverResolveByNameFunc) nextHook() func(context.Context, string) (*Teammate, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *TeammateResolverResolveByNameFunc) appendCall(r0 TeammateResolverResolveByNameFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of TeammateResolverResolveByNameFuncCall
// objects describing the invocations of this function.
func (f *TeammateResolverResolveByNameFunc) History() []TeammateResolverResolveByNameFuncCall {
	f.mutex.Lock()
	history := make([]TeammateResolverResolveByNameFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// TeammateResolverResolveByNameFuncCall is an object that describes an
// invocation of method ResolveByName on an instance of
// MockTeammateResolver.
type TeammateResolverResolveByNameFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *Teammate
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c TeammateResolverResolveByNameFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c TeammateResolverResolveByNameFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
