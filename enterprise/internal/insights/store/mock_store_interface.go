// Code generated by go-mockgen 1.3.1; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package store

import (
	"context"
	"sync"
)

// MockInterface is a mock implementation of the Interface interface (from
// the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store)
// used for unit testing.
type MockInterface struct {
	// CountDataFunc is an instance of a mock function object controlling
	// the behavior of the method CountData.
	CountDataFunc *InterfaceCountDataFunc
	// RecordSeriesPointFunc is an instance of a mock function object
	// controlling the behavior of the method RecordSeriesPoint.
	RecordSeriesPointFunc *InterfaceRecordSeriesPointFunc
	// RecordSeriesPointsFunc is an instance of a mock function object
	// controlling the behavior of the method RecordSeriesPoints.
	RecordSeriesPointsFunc *InterfaceRecordSeriesPointsFunc
	// SeriesPointsFunc is an instance of a mock function object controlling
	// the behavior of the method SeriesPoints.
	SeriesPointsFunc *InterfaceSeriesPointsFunc
}

// NewMockInterface creates a new mock of the Interface interface. All
// methods return zero values for all results, unless overwritten.
func NewMockInterface() *MockInterface {
	return &MockInterface{
		CountDataFunc: &InterfaceCountDataFunc{
			defaultHook: func(context.Context, CountDataOpts) (r0 int, r1 error) {
				return
			},
		},
		RecordSeriesPointFunc: &InterfaceRecordSeriesPointFunc{
			defaultHook: func(context.Context, RecordSeriesPointArgs) (r0 error) {
				return
			},
		},
		RecordSeriesPointsFunc: &InterfaceRecordSeriesPointsFunc{
			defaultHook: func(context.Context, []RecordSeriesPointArgs) (r0 error) {
				return
			},
		},
		SeriesPointsFunc: &InterfaceSeriesPointsFunc{
			defaultHook: func(context.Context, SeriesPointsOpts) (r0 []SeriesPoint, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockInterface creates a new mock of the Interface interface. All
// methods panic on invocation, unless overwritten.
func NewStrictMockInterface() *MockInterface {
	return &MockInterface{
		CountDataFunc: &InterfaceCountDataFunc{
			defaultHook: func(context.Context, CountDataOpts) (int, error) {
				panic("unexpected invocation of MockInterface.CountData")
			},
		},
		RecordSeriesPointFunc: &InterfaceRecordSeriesPointFunc{
			defaultHook: func(context.Context, RecordSeriesPointArgs) error {
				panic("unexpected invocation of MockInterface.RecordSeriesPoint")
			},
		},
		RecordSeriesPointsFunc: &InterfaceRecordSeriesPointsFunc{
			defaultHook: func(context.Context, []RecordSeriesPointArgs) error {
				panic("unexpected invocation of MockInterface.RecordSeriesPoints")
			},
		},
		SeriesPointsFunc: &InterfaceSeriesPointsFunc{
			defaultHook: func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
				panic("unexpected invocation of MockInterface.SeriesPoints")
			},
		},
	}
}

// NewMockInterfaceFrom creates a new mock of the MockInterface interface.
// All methods delegate to the given implementation, unless overwritten.
func NewMockInterfaceFrom(i Interface) *MockInterface {
	return &MockInterface{
		CountDataFunc: &InterfaceCountDataFunc{
			defaultHook: i.CountData,
		},
		RecordSeriesPointFunc: &InterfaceRecordSeriesPointFunc{
			defaultHook: i.RecordSeriesPoint,
		},
		RecordSeriesPointsFunc: &InterfaceRecordSeriesPointsFunc{
			defaultHook: i.RecordSeriesPoints,
		},
		SeriesPointsFunc: &InterfaceSeriesPointsFunc{
			defaultHook: i.SeriesPoints,
		},
	}
}

// InterfaceCountDataFunc describes the behavior when the CountData method
// of the parent MockInterface instance is invoked.
type InterfaceCountDataFunc struct {
	defaultHook func(context.Context, CountDataOpts) (int, error)
	hooks       []func(context.Context, CountDataOpts) (int, error)
	history     []InterfaceCountDataFuncCall
	mutex       sync.Mutex
}

// CountData delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockInterface) CountData(v0 context.Context, v1 CountDataOpts) (int, error) {
	r0, r1 := m.CountDataFunc.nextHook()(v0, v1)
	m.CountDataFunc.appendCall(InterfaceCountDataFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the CountData method of
// the parent MockInterface instance is invoked and the hook queue is empty.
func (f *InterfaceCountDataFunc) SetDefaultHook(hook func(context.Context, CountDataOpts) (int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CountData method of the parent MockInterface instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *InterfaceCountDataFunc) PushHook(hook func(context.Context, CountDataOpts) (int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *InterfaceCountDataFunc) SetDefaultReturn(r0 int, r1 error) {
	f.SetDefaultHook(func(context.Context, CountDataOpts) (int, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *InterfaceCountDataFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, CountDataOpts) (int, error) {
		return r0, r1
	})
}

func (f *InterfaceCountDataFunc) nextHook() func(context.Context, CountDataOpts) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfaceCountDataFunc) appendCall(r0 InterfaceCountDataFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of InterfaceCountDataFuncCall objects
// describing the invocations of this function.
func (f *InterfaceCountDataFunc) History() []InterfaceCountDataFuncCall {
	f.mutex.Lock()
	history := make([]InterfaceCountDataFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfaceCountDataFuncCall is an object that describes an invocation of
// method CountData on an instance of MockInterface.
type InterfaceCountDataFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 CountDataOpts
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c InterfaceCountDataFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c InterfaceCountDataFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// InterfaceRecordSeriesPointFunc describes the behavior when the
// RecordSeriesPoint method of the parent MockInterface instance is invoked.
type InterfaceRecordSeriesPointFunc struct {
	defaultHook func(context.Context, RecordSeriesPointArgs) error
	hooks       []func(context.Context, RecordSeriesPointArgs) error
	history     []InterfaceRecordSeriesPointFuncCall
	mutex       sync.Mutex
}

// RecordSeriesPoint delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockInterface) RecordSeriesPoint(v0 context.Context, v1 RecordSeriesPointArgs) error {
	r0 := m.RecordSeriesPointFunc.nextHook()(v0, v1)
	m.RecordSeriesPointFunc.appendCall(InterfaceRecordSeriesPointFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the RecordSeriesPoint
// method of the parent MockInterface instance is invoked and the hook queue
// is empty.
func (f *InterfaceRecordSeriesPointFunc) SetDefaultHook(hook func(context.Context, RecordSeriesPointArgs) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RecordSeriesPoint method of the parent MockInterface instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *InterfaceRecordSeriesPointFunc) PushHook(hook func(context.Context, RecordSeriesPointArgs) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *InterfaceRecordSeriesPointFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, RecordSeriesPointArgs) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *InterfaceRecordSeriesPointFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, RecordSeriesPointArgs) error {
		return r0
	})
}

func (f *InterfaceRecordSeriesPointFunc) nextHook() func(context.Context, RecordSeriesPointArgs) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfaceRecordSeriesPointFunc) appendCall(r0 InterfaceRecordSeriesPointFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of InterfaceRecordSeriesPointFuncCall objects
// describing the invocations of this function.
func (f *InterfaceRecordSeriesPointFunc) History() []InterfaceRecordSeriesPointFuncCall {
	f.mutex.Lock()
	history := make([]InterfaceRecordSeriesPointFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfaceRecordSeriesPointFuncCall is an object that describes an
// invocation of method RecordSeriesPoint on an instance of MockInterface.
type InterfaceRecordSeriesPointFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 RecordSeriesPointArgs
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c InterfaceRecordSeriesPointFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c InterfaceRecordSeriesPointFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// InterfaceRecordSeriesPointsFunc describes the behavior when the
// RecordSeriesPoints method of the parent MockInterface instance is
// invoked.
type InterfaceRecordSeriesPointsFunc struct {
	defaultHook func(context.Context, []RecordSeriesPointArgs) error
	hooks       []func(context.Context, []RecordSeriesPointArgs) error
	history     []InterfaceRecordSeriesPointsFuncCall
	mutex       sync.Mutex
}

// RecordSeriesPoints delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockInterface) RecordSeriesPoints(v0 context.Context, v1 []RecordSeriesPointArgs) error {
	r0 := m.RecordSeriesPointsFunc.nextHook()(v0, v1)
	m.RecordSeriesPointsFunc.appendCall(InterfaceRecordSeriesPointsFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the RecordSeriesPoints
// method of the parent MockInterface instance is invoked and the hook queue
// is empty.
func (f *InterfaceRecordSeriesPointsFunc) SetDefaultHook(hook func(context.Context, []RecordSeriesPointArgs) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RecordSeriesPoints method of the parent MockInterface instance invokes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *InterfaceRecordSeriesPointsFunc) PushHook(hook func(context.Context, []RecordSeriesPointArgs) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *InterfaceRecordSeriesPointsFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, []RecordSeriesPointArgs) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *InterfaceRecordSeriesPointsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []RecordSeriesPointArgs) error {
		return r0
	})
}

func (f *InterfaceRecordSeriesPointsFunc) nextHook() func(context.Context, []RecordSeriesPointArgs) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfaceRecordSeriesPointsFunc) appendCall(r0 InterfaceRecordSeriesPointsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of InterfaceRecordSeriesPointsFuncCall objects
// describing the invocations of this function.
func (f *InterfaceRecordSeriesPointsFunc) History() []InterfaceRecordSeriesPointsFuncCall {
	f.mutex.Lock()
	history := make([]InterfaceRecordSeriesPointsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfaceRecordSeriesPointsFuncCall is an object that describes an
// invocation of method RecordSeriesPoints on an instance of MockInterface.
type InterfaceRecordSeriesPointsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 []RecordSeriesPointArgs
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c InterfaceRecordSeriesPointsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c InterfaceRecordSeriesPointsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// InterfaceSeriesPointsFunc describes the behavior when the SeriesPoints
// method of the parent MockInterface instance is invoked.
type InterfaceSeriesPointsFunc struct {
	defaultHook func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)
	hooks       []func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)
	history     []InterfaceSeriesPointsFuncCall
	mutex       sync.Mutex
}

// SeriesPoints delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockInterface) SeriesPoints(v0 context.Context, v1 SeriesPointsOpts) ([]SeriesPoint, error) {
	r0, r1 := m.SeriesPointsFunc.nextHook()(v0, v1)
	m.SeriesPointsFunc.appendCall(InterfaceSeriesPointsFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the SeriesPoints method
// of the parent MockInterface instance is invoked and the hook queue is
// empty.
func (f *InterfaceSeriesPointsFunc) SetDefaultHook(hook func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// SeriesPoints method of the parent MockInterface instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *InterfaceSeriesPointsFunc) PushHook(hook func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *InterfaceSeriesPointsFunc) SetDefaultReturn(r0 []SeriesPoint, r1 error) {
	f.SetDefaultHook(func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *InterfaceSeriesPointsFunc) PushReturn(r0 []SeriesPoint, r1 error) {
	f.PushHook(func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
		return r0, r1
	})
}

func (f *InterfaceSeriesPointsFunc) nextHook() func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfaceSeriesPointsFunc) appendCall(r0 InterfaceSeriesPointsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of InterfaceSeriesPointsFuncCall objects
// describing the invocations of this function.
func (f *InterfaceSeriesPointsFunc) History() []InterfaceSeriesPointsFuncCall {
	f.mutex.Lock()
	history := make([]InterfaceSeriesPointsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfaceSeriesPointsFuncCall is an object that describes an invocation
// of method SeriesPoints on an instance of MockInterface.
type InterfaceSeriesPointsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 SeriesPointsOpts
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []SeriesPoint
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c InterfaceSeriesPointsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c InterfaceSeriesPointsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}
