package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockUsers struct {
	GetByID                                         func(ctx context.Context, id int32) (*types.User, error)
	GetByUsername                                   func(ctx context.Context, username string) (*types.User, error)
	GetByCurrentAuthUser                            func(ctx context.Context) (*types.User, error)
	Count                                           func(ctx context.Context, opt *UsersListOptions) (int, error)
	InvalidateSessionsByID                          func(ctx context.Context, id int32) error
	HasTag                                          func(ctx context.Context, userID int32, tag string) (bool, error)
	Tags                                            func(ctx context.Context, userID int32) (map[string]bool, error)
	RandomizePasswordAndClearPasswordResetRateLimit func(ctx context.Context, userID int32) error
	RenewPasswordResetCode                          func(ctx context.Context, id int32) (string, error)
}

func (s *MockUsers) MockGetByID_Return(t *testing.T, returns *types.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
