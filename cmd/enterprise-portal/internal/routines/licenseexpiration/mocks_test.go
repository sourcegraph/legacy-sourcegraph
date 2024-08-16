// Code generated by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package licenseexpiration

import (
	"context"
	"sync"
	"time"

	subscriptions "github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	slack "github.com/sourcegraph/sourcegraph/internal/slack"
)

// MockStore is a mock implementation of the Store interface (from the
// package
// github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/routines/licenseexpiration)
// used for unit testing.
type MockStore struct {
	// EnvFunc is an instance of a mock function object controlling the
	// behavior of the method Env.
	EnvFunc *StoreEnvFunc
	// GetActiveLicenseFunc is an instance of a mock function object
	// controlling the behavior of the method GetActiveLicense.
	GetActiveLicenseFunc *StoreGetActiveLicenseFunc
	// ListSubscriptionsFunc is an instance of a mock function object
	// controlling the behavior of the method ListSubscriptions.
	ListSubscriptionsFunc *StoreListSubscriptionsFunc
	// NowFunc is an instance of a mock function object controlling the
	// behavior of the method Now.
	NowFunc *StoreNowFunc
	// PostToSlackFunc is an instance of a mock function object controlling
	// the behavior of the method PostToSlack.
	PostToSlackFunc *StorePostToSlackFunc
	// TryAcquireJobFunc is an instance of a mock function object
	// controlling the behavior of the method TryAcquireJob.
	TryAcquireJobFunc *StoreTryAcquireJobFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		EnvFunc: &StoreEnvFunc{
			defaultHook: func() (r0 string) {
				return
			},
		},
		GetActiveLicenseFunc: &StoreGetActiveLicenseFunc{
			defaultHook: func(context.Context, string) (r0 *subscriptions.LicenseWithConditions, r1 error) {
				return
			},
		},
		ListSubscriptionsFunc: &StoreListSubscriptionsFunc{
			defaultHook: func(context.Context) (r0 []*subscriptions.SubscriptionWithConditions, r1 error) {
				return
			},
		},
		NowFunc: &StoreNowFunc{
			defaultHook: func() (r0 time.Time) {
				return
			},
		},
		PostToSlackFunc: &StorePostToSlackFunc{
			defaultHook: func(context.Context, *slack.Payload) (r0 error) {
				return
			},
		},
		TryAcquireJobFunc: &StoreTryAcquireJobFunc{
			defaultHook: func(context.Context) (r0 bool, r1 func(), r2 error) {
				return
			},
		},
	}
}

// NewStrictMockStore creates a new mock of the Store interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		EnvFunc: &StoreEnvFunc{
			defaultHook: func() string {
				panic("unexpected invocation of MockStore.Env")
			},
		},
		GetActiveLicenseFunc: &StoreGetActiveLicenseFunc{
			defaultHook: func(context.Context, string) (*subscriptions.LicenseWithConditions, error) {
				panic("unexpected invocation of MockStore.GetActiveLicense")
			},
		},
		ListSubscriptionsFunc: &StoreListSubscriptionsFunc{
			defaultHook: func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error) {
				panic("unexpected invocation of MockStore.ListSubscriptions")
			},
		},
		NowFunc: &StoreNowFunc{
			defaultHook: func() time.Time {
				panic("unexpected invocation of MockStore.Now")
			},
		},
		PostToSlackFunc: &StorePostToSlackFunc{
			defaultHook: func(context.Context, *slack.Payload) error {
				panic("unexpected invocation of MockStore.PostToSlack")
			},
		},
		TryAcquireJobFunc: &StoreTryAcquireJobFunc{
			defaultHook: func(context.Context) (bool, func(), error) {
				panic("unexpected invocation of MockStore.TryAcquireJob")
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		EnvFunc: &StoreEnvFunc{
			defaultHook: i.Env,
		},
		GetActiveLicenseFunc: &StoreGetActiveLicenseFunc{
			defaultHook: i.GetActiveLicense,
		},
		ListSubscriptionsFunc: &StoreListSubscriptionsFunc{
			defaultHook: i.ListSubscriptions,
		},
		NowFunc: &StoreNowFunc{
			defaultHook: i.Now,
		},
		PostToSlackFunc: &StorePostToSlackFunc{
			defaultHook: i.PostToSlack,
		},
		TryAcquireJobFunc: &StoreTryAcquireJobFunc{
			defaultHook: i.TryAcquireJob,
		},
	}
}

// StoreEnvFunc describes the behavior when the Env method of the parent
// MockStore instance is invoked.
type StoreEnvFunc struct {
	defaultHook func() string
	hooks       []func() string
	history     []StoreEnvFuncCall
	mutex       sync.Mutex
}

// Env delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Env() string {
	r0 := m.EnvFunc.nextHook()()
	m.EnvFunc.appendCall(StoreEnvFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Env method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreEnvFunc) SetDefaultHook(hook func() string) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Env method of the parent MockStore instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *StoreEnvFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreEnvFunc) SetDefaultReturn(r0 string) {
	f.SetDefaultHook(func() string {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreEnvFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *StoreEnvFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreEnvFunc) appendCall(r0 StoreEnvFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreEnvFuncCall objects describing the
// invocations of this function.
func (f *StoreEnvFunc) History() []StoreEnvFuncCall {
	f.mutex.Lock()
	history := make([]StoreEnvFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreEnvFuncCall is an object that describes an invocation of method Env
// on an instance of MockStore.
type StoreEnvFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreEnvFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreEnvFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreGetActiveLicenseFunc describes the behavior when the
// GetActiveLicense method of the parent MockStore instance is invoked.
type StoreGetActiveLicenseFunc struct {
	defaultHook func(context.Context, string) (*subscriptions.LicenseWithConditions, error)
	hooks       []func(context.Context, string) (*subscriptions.LicenseWithConditions, error)
	history     []StoreGetActiveLicenseFuncCall
	mutex       sync.Mutex
}

// GetActiveLicense delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) GetActiveLicense(v0 context.Context, v1 string) (*subscriptions.LicenseWithConditions, error) {
	r0, r1 := m.GetActiveLicenseFunc.nextHook()(v0, v1)
	m.GetActiveLicenseFunc.appendCall(StoreGetActiveLicenseFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the GetActiveLicense
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreGetActiveLicenseFunc) SetDefaultHook(hook func(context.Context, string) (*subscriptions.LicenseWithConditions, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// GetActiveLicense method of the parent MockStore instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreGetActiveLicenseFunc) PushHook(hook func(context.Context, string) (*subscriptions.LicenseWithConditions, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreGetActiveLicenseFunc) SetDefaultReturn(r0 *subscriptions.LicenseWithConditions, r1 error) {
	f.SetDefaultHook(func(context.Context, string) (*subscriptions.LicenseWithConditions, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreGetActiveLicenseFunc) PushReturn(r0 *subscriptions.LicenseWithConditions, r1 error) {
	f.PushHook(func(context.Context, string) (*subscriptions.LicenseWithConditions, error) {
		return r0, r1
	})
}

func (f *StoreGetActiveLicenseFunc) nextHook() func(context.Context, string) (*subscriptions.LicenseWithConditions, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetActiveLicenseFunc) appendCall(r0 StoreGetActiveLicenseFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreGetActiveLicenseFuncCall objects
// describing the invocations of this function.
func (f *StoreGetActiveLicenseFunc) History() []StoreGetActiveLicenseFuncCall {
	f.mutex.Lock()
	history := make([]StoreGetActiveLicenseFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetActiveLicenseFuncCall is an object that describes an invocation
// of method GetActiveLicense on an instance of MockStore.
type StoreGetActiveLicenseFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *subscriptions.LicenseWithConditions
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreGetActiveLicenseFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreGetActiveLicenseFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreListSubscriptionsFunc describes the behavior when the
// ListSubscriptions method of the parent MockStore instance is invoked.
type StoreListSubscriptionsFunc struct {
	defaultHook func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error)
	hooks       []func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error)
	history     []StoreListSubscriptionsFuncCall
	mutex       sync.Mutex
}

// ListSubscriptions delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) ListSubscriptions(v0 context.Context) ([]*subscriptions.SubscriptionWithConditions, error) {
	r0, r1 := m.ListSubscriptionsFunc.nextHook()(v0)
	m.ListSubscriptionsFunc.appendCall(StoreListSubscriptionsFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the ListSubscriptions
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreListSubscriptionsFunc) SetDefaultHook(hook func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// ListSubscriptions method of the parent MockStore instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *StoreListSubscriptionsFunc) PushHook(hook func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreListSubscriptionsFunc) SetDefaultReturn(r0 []*subscriptions.SubscriptionWithConditions, r1 error) {
	f.SetDefaultHook(func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreListSubscriptionsFunc) PushReturn(r0 []*subscriptions.SubscriptionWithConditions, r1 error) {
	f.PushHook(func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error) {
		return r0, r1
	})
}

func (f *StoreListSubscriptionsFunc) nextHook() func(context.Context) ([]*subscriptions.SubscriptionWithConditions, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreListSubscriptionsFunc) appendCall(r0 StoreListSubscriptionsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreListSubscriptionsFuncCall objects
// describing the invocations of this function.
func (f *StoreListSubscriptionsFunc) History() []StoreListSubscriptionsFuncCall {
	f.mutex.Lock()
	history := make([]StoreListSubscriptionsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreListSubscriptionsFuncCall is an object that describes an invocation
// of method ListSubscriptions on an instance of MockStore.
type StoreListSubscriptionsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []*subscriptions.SubscriptionWithConditions
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreListSubscriptionsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreListSubscriptionsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreNowFunc describes the behavior when the Now method of the parent
// MockStore instance is invoked.
type StoreNowFunc struct {
	defaultHook func() time.Time
	hooks       []func() time.Time
	history     []StoreNowFuncCall
	mutex       sync.Mutex
}

// Now delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Now() time.Time {
	r0 := m.NowFunc.nextHook()()
	m.NowFunc.appendCall(StoreNowFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Now method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreNowFunc) SetDefaultHook(hook func() time.Time) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Now method of the parent MockStore instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *StoreNowFunc) PushHook(hook func() time.Time) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreNowFunc) SetDefaultReturn(r0 time.Time) {
	f.SetDefaultHook(func() time.Time {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreNowFunc) PushReturn(r0 time.Time) {
	f.PushHook(func() time.Time {
		return r0
	})
}

func (f *StoreNowFunc) nextHook() func() time.Time {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreNowFunc) appendCall(r0 StoreNowFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreNowFuncCall objects describing the
// invocations of this function.
func (f *StoreNowFunc) History() []StoreNowFuncCall {
	f.mutex.Lock()
	history := make([]StoreNowFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreNowFuncCall is an object that describes an invocation of method Now
// on an instance of MockStore.
type StoreNowFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 time.Time
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreNowFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreNowFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StorePostToSlackFunc describes the behavior when the PostToSlack method
// of the parent MockStore instance is invoked.
type StorePostToSlackFunc struct {
	defaultHook func(context.Context, *slack.Payload) error
	hooks       []func(context.Context, *slack.Payload) error
	history     []StorePostToSlackFuncCall
	mutex       sync.Mutex
}

// PostToSlack delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) PostToSlack(v0 context.Context, v1 *slack.Payload) error {
	r0 := m.PostToSlackFunc.nextHook()(v0, v1)
	m.PostToSlackFunc.appendCall(StorePostToSlackFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the PostToSlack method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StorePostToSlackFunc) SetDefaultHook(hook func(context.Context, *slack.Payload) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// PostToSlack method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StorePostToSlackFunc) PushHook(hook func(context.Context, *slack.Payload) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StorePostToSlackFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, *slack.Payload) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StorePostToSlackFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *slack.Payload) error {
		return r0
	})
}

func (f *StorePostToSlackFunc) nextHook() func(context.Context, *slack.Payload) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StorePostToSlackFunc) appendCall(r0 StorePostToSlackFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StorePostToSlackFuncCall objects describing
// the invocations of this function.
func (f *StorePostToSlackFunc) History() []StorePostToSlackFuncCall {
	f.mutex.Lock()
	history := make([]StorePostToSlackFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StorePostToSlackFuncCall is an object that describes an invocation of
// method PostToSlack on an instance of MockStore.
type StorePostToSlackFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 *slack.Payload
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StorePostToSlackFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StorePostToSlackFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreTryAcquireJobFunc describes the behavior when the TryAcquireJob
// method of the parent MockStore instance is invoked.
type StoreTryAcquireJobFunc struct {
	defaultHook func(context.Context) (bool, func(), error)
	hooks       []func(context.Context) (bool, func(), error)
	history     []StoreTryAcquireJobFuncCall
	mutex       sync.Mutex
}

// TryAcquireJob delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) TryAcquireJob(v0 context.Context) (bool, func(), error) {
	r0, r1, r2 := m.TryAcquireJobFunc.nextHook()(v0)
	m.TryAcquireJobFunc.appendCall(StoreTryAcquireJobFuncCall{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the TryAcquireJob method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreTryAcquireJobFunc) SetDefaultHook(hook func(context.Context) (bool, func(), error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// TryAcquireJob method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreTryAcquireJobFunc) PushHook(hook func(context.Context) (bool, func(), error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *StoreTryAcquireJobFunc) SetDefaultReturn(r0 bool, r1 func(), r2 error) {
	f.SetDefaultHook(func(context.Context) (bool, func(), error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *StoreTryAcquireJobFunc) PushReturn(r0 bool, r1 func(), r2 error) {
	f.PushHook(func(context.Context) (bool, func(), error) {
		return r0, r1, r2
	})
}

func (f *StoreTryAcquireJobFunc) nextHook() func(context.Context) (bool, func(), error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTryAcquireJobFunc) appendCall(r0 StoreTryAcquireJobFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreTryAcquireJobFuncCall objects
// describing the invocations of this function.
func (f *StoreTryAcquireJobFunc) History() []StoreTryAcquireJobFuncCall {
	f.mutex.Lock()
	history := make([]StoreTryAcquireJobFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTryAcquireJobFuncCall is an object that describes an invocation of
// method TryAcquireJob on an instance of MockStore.
type StoreTryAcquireJobFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 func()
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreTryAcquireJobFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreTryAcquireJobFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}
