package search

import (
	"context"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/tenant"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestConvertToGRPCError(t *testing.T) {
	expiredContext, done := context.WithDeadline(tenant.TestContext(), time.Now().Add(-time.Minute))
	t.Cleanup(func() {
		done()
	})

	testCases := []struct {
		name string

		ctx context.Context
		err error

		want error
	}{
		{
			name: "nil error",

			ctx: tenant.TestContext(),
			err: nil,

			want: nil,
		},
		{
			name: "existing status error",

			ctx: tenant.TestContext(),
			err: status.Error(codes.InvalidArgument, "invalid"),

			want: status.Error(codes.InvalidArgument, "invalid"),
		},
		{
			name: "context error",

			ctx: expiredContext,
			err: errors.New("some other error"),

			want: status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error()),
		},
		{
			name: "unknown error",

			ctx: tenant.TestContext(),
			err: errors.New("unknown"),

			want: status.Error(codes.Unknown, "unknown"),
		},
		{
			name: "killed error",

			ctx: tenant.TestContext(),
			err: errors.New("failed to wait for executing comby command: signal: killed"),

			want: status.Error(codes.Aborted, "failed to wait for executing comby command: signal: killed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertToGRPCError(tc.ctx, tc.err)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("convertToGRPCError() = %v, want %v", got, tc.want)
			}
		})
	}
}
