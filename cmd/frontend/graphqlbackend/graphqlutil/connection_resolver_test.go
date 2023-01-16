package graphqlutil

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const testTotalCount = int32(10)

type testConnectionNode struct {
	id int
}

func (n testConnectionNode) ID() graphql.ID {
	return graphql.ID(fmt.Sprintf("%d", n.id))
}

type testConnectionStore struct {
	t                      *testing.T
	expectedPaginationArgs *database.PaginationArgs
	ComputeTotalCalled     int
	ComputeNodesCalled     int
}

func (s *testConnectionStore) testPaginationArgs(args *database.PaginationArgs) {
	if s.expectedPaginationArgs == nil {
		return
	}

	if diff := cmp.Diff(s.expectedPaginationArgs, args); diff != "" {
		s.t.Fatal(diff)
	}
}

func (s *testConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	s.ComputeTotalCalled = s.ComputeTotalCalled + 1
	total := testTotalCount

	return &total, nil
}

func (s *testConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*testConnectionNode, error) {
	s.ComputeNodesCalled = s.ComputeNodesCalled + 1
	s.testPaginationArgs(args)

	if args.First != nil {
		return []*testConnectionNode{{id: 0}, {id: 1}}, nil
	}

	// Return in the reverse order because a SQL query will need to ORDER BY DESC to support
	// args.Last.
	if args.Last != nil {
		return []*testConnectionNode{{id: 1}, {id: 0}}, nil
	}

	return []*testConnectionNode{}, errors.New("unexpected: either of First or Last must be set")
}

func (*testConnectionStore) MarshalCursor(n *testConnectionNode) (*string, error) {
	cursor := string(n.ID())

	return &cursor, nil
}

func (*testConnectionStore) UnmarshalCursor(cursor string) (*int, error) {
	id, err := strconv.Atoi(cursor)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func newInt(n int) *int {
	return &n
}

func newInt32(n int) *int32 {
	num := int32(n)

	return &num
}

func withFirstCA(first int, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.First = newInt32(first)

	return a
}

func withLastCA(last int, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.Last = newInt32(last)

	return a
}

func withAfterCA(after string, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.After = &after

	return a
}

func withBeforeCA(before string, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.Before = &before

	return a
}

func withFirstPA(first int, a *database.PaginationArgs) *database.PaginationArgs {
	a.First = &first

	return a
}

func withLastPA(last int, a *database.PaginationArgs) *database.PaginationArgs {
	a.Last = &last

	return a
}

func withAfterPA(after int, a *database.PaginationArgs) *database.PaginationArgs {
	a.After = &after

	return a
}

func withBeforePA(before int, a *database.PaginationArgs) *database.PaginationArgs {
	a.Before = &before

	return a
}

func TestConnectionTotalCount(t *testing.T) {
	ctx := context.Background()
	store := &testConnectionStore{t: t}
	resolver, err := NewConnectionResolver[testConnectionNode](store, withFirstCA(1, &ConnectionResolverArgs{}), nil)
	if err != nil {
		t.Fatal(err)
	}

	count, err := resolver.TotalCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != testTotalCount {
		t.Fatalf("wrong total count. want=%d, have=%d", testTotalCount, count)
	}

	resolver.TotalCount(ctx)
	if store.ComputeTotalCalled != 1 {
		t.Fatalf("wrong compute total called count. want=%d, have=%d", 1, store.ComputeTotalCalled)
	}
}

func testResolverNodesResponse(t *testing.T, resolver *ConnectionResolver[testConnectionNode], store *testConnectionStore, count int) {
	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(count, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	resolver.Nodes(ctx)
	if store.ComputeNodesCalled != 1 {
		t.Fatalf("wrong compute nodes called count. want=%d, have=%d", 1, store.ComputeNodesCalled)
	}
}

func TestConnectionNodes(t *testing.T) {
	for _, test := range []struct {
		name           string
		connectionArgs *ConnectionResolverArgs

		wantPaginationArgs *database.PaginationArgs
		wantNodes          int
	}{
		{
			name:               "default",
			connectionArgs:     withFirstCA(5, &ConnectionResolverArgs{}),
			wantPaginationArgs: withFirstPA(6, &database.PaginationArgs{}),
			wantNodes:          2,
		},
		{
			name:               "last arg",
			wantPaginationArgs: withLastPA(6, &database.PaginationArgs{}),
			connectionArgs:     withLastCA(5, &ConnectionResolverArgs{}),
			wantNodes:          2,
		},
		{
			name:               "after arg",
			wantPaginationArgs: withAfterPA(0, withFirstPA(6, &database.PaginationArgs{})),
			connectionArgs:     withAfterCA("0", withFirstCA(5, &ConnectionResolverArgs{})),
			wantNodes:          2,
		},
		{
			name:               "before arg",
			wantPaginationArgs: withBeforePA(0, withLastPA(6, &database.PaginationArgs{})),
			connectionArgs:     withBeforeCA("0", withLastCA(5, &ConnectionResolverArgs{})),
			wantNodes:          2,
		},
		{
			name:               "with limit",
			wantPaginationArgs: withBeforePA(0, withLastPA(2, &database.PaginationArgs{})),
			connectionArgs:     withBeforeCA("0", withLastCA(1, &ConnectionResolverArgs{})),
			wantNodes:          1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &testConnectionStore{t: t, expectedPaginationArgs: test.wantPaginationArgs}
			resolver, err := NewConnectionResolver[testConnectionNode](store, test.connectionArgs, nil)
			if err != nil {
				t.Fatal(err)
			}

			testResolverNodesResponse(t, resolver, store, test.wantNodes)
		})
	}
}

type pageInfoResponse struct {
	startCursor     string
	endCursor       string
	hasNextPage     bool
	hasPreviousPage bool
}

func testResolverPageInfoResponse(t *testing.T, resolver *ConnectionResolver[testConnectionNode], store *testConnectionStore, expectedResponse *pageInfoResponse) {
	ctx := context.Background()
	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expectedResponse.startCursor, *startCursor); diff != "" {
		t.Errorf("mismatched startCursor in response: %v", diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expectedResponse.endCursor, *endCursor); diff != "" {
		t.Errorf("mismatched endCursor in response: %v", diff)
	}

	if expectedResponse.hasNextPage != pageInfo.HasNextPage() {
		t.Errorf("hasNextPage should be %v, but is %v", expectedResponse.hasNextPage, pageInfo.HasNextPage())
	}
	if expectedResponse.hasPreviousPage != pageInfo.HasPreviousPage() {
		t.Errorf("hasPreviousPage should be %v, but is %v", expectedResponse.hasPreviousPage, pageInfo.HasPreviousPage())
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfo(t *testing.T) {
	for _, test := range []struct {
		name string
		args *ConnectionResolverArgs
		want *pageInfoResponse
	}{
		{
			name: "default",
			args: withFirstCA(20, &ConnectionResolverArgs{}),
			want: &pageInfoResponse{startCursor: "0", endCursor: "1", hasNextPage: false, hasPreviousPage: false},
		},
		{
			name: "first page",
			args: withFirstCA(1, &ConnectionResolverArgs{}),
			want: &pageInfoResponse{startCursor: "0", endCursor: "0", hasNextPage: true, hasPreviousPage: false},
		},
		{
			name: "second page",
			args: withAfterCA("0", withFirstCA(1, &ConnectionResolverArgs{})),
			want: &pageInfoResponse{startCursor: "0", endCursor: "0", hasNextPage: true, hasPreviousPage: true},
		},
		{
			name: "backward first page",
			args: withBeforeCA("0", withLastCA(1, &ConnectionResolverArgs{})),
			want: &pageInfoResponse{startCursor: "1", endCursor: "1", hasNextPage: true, hasPreviousPage: true},
		},
		{
			name: "backward first page without cursor",
			args: withLastCA(1, &ConnectionResolverArgs{}),
			want: &pageInfoResponse{startCursor: "1", endCursor: "1", hasNextPage: false, hasPreviousPage: true},
		},
		{
			name: "backward last page",
			args: &ConnectionResolverArgs{
				Last:   newInt32(20),
				Before: stringPtr("0"),
			},
			want: &pageInfoResponse{startCursor: "1", endCursor: "0", hasNextPage: true, hasPreviousPage: false},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &testConnectionStore{t: t}
			resolver, err := NewConnectionResolver[testConnectionNode](store, test.args, nil)
			if err != nil {
				t.Fatal(err)
			}
			testResolverPageInfoResponse(t, resolver, store, test.want)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
