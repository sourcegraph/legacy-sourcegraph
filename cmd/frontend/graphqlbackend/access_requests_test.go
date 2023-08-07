package graphqlbackend

import (
	"context"
	"testing"
	"time"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/accessrequests"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAccessRequestNode(t *testing.T) {
	mockAccessRequest := &types.AccessRequest{
		ID:             1,
		Email:          "a1@example.com",
		Name:           "a1",
		CreatedAt:      time.Now(),
		AdditionalInfo: "af1",
		Status:         types.AccessRequestStatusPending,
	}
	db := database.NewMockDB()

	mockDBClient := database.NewMockDBClient()
	db.ClientFunc.SetDefaultReturn(mockDBClient)
	mockDBClient.Mock(&accessrequests.GetByID{}, mockAccessRequest, nil)

	userStore := database.NewMockUserStore()
	db.UsersFunc.SetDefaultReturn(userStore)
	userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `
		query AccessRequestID($id: ID!){
			node(id: $id) {
				__typename
				... on AccessRequest {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"node": {
				"__typename": "AccessRequest",
				"name": "a1"
			}
		}`,
		Variables: map[string]any{
			"id": string(marshalAccessRequestID(mockAccessRequest.ID)),
		},
	})
}

func TestAccessRequestsQuery(t *testing.T) {
	const accessRequestsQuery = `
	query GetAccessRequests($first: Int, $after: String, $before: String, $last: Int) {
		accessRequests(first: $first, after: $after, before: $before, last: $last) {
			nodes {
				id
				name
				email
				status
				createdAt
				additionalInfo
			}
			totalCount
			pageInfo {
				hasNextPage
				hasPreviousPage
				startCursor
				endCursor
			}
		}
	}`

	db := database.NewMockDB()

	userStore := database.NewMockUserStore()
	db.UsersFunc.SetDefaultReturn(userStore)

	mockDBClient := database.NewMockDBClient()
	db.ClientFunc.SetDefaultReturn(mockDBClient)

	t.Parallel()

	t.Run("non-admin user", func(t *testing.T) {
		userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Context:        ctx,
			Query:          accessRequestsQuery,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"accessRequests"},
					Message:       auth.ErrMustBeSiteAdmin.Error(),
					ResolverError: auth.ErrMustBeSiteAdmin,
				},
			},
			Variables: map[string]any{
				"first": 10,
			},
		})
	})

	t.Run("admin user", func(t *testing.T) {
		createdAtTime, _ := time.Parse(time.RFC3339, "2023-02-24T14:48:30Z")
		mockAccessRequests := []*types.AccessRequest{
			{ID: 1, Email: "a1@example.com", Name: "a1", CreatedAt: createdAtTime, AdditionalInfo: "af1", Status: types.AccessRequestStatusPending},
			{ID: 2, Email: "a2@example.com", Name: "a2", CreatedAt: createdAtTime, Status: types.AccessRequestStatusApproved},
			{ID: 3, Email: "a3@example.com", Name: "a3", CreatedAt: createdAtTime, Status: types.AccessRequestStatusRejected},
		}

		mockDBClient.Mock(&accessrequests.List{}, mockAccessRequests, nil)
		mockDBClient.Mock(&accessrequests.Count{}, len(mockAccessRequests), nil)
		userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		RunTest(t, &Test{
			Schema:  mustParseGraphQLSchema(t, db),
			Context: ctx,
			Query:   accessRequestsQuery,
			ExpectedResult: `{
				"accessRequests": {
					"nodes": [
						{
							"id": "QWNjZXNzUmVxdWVzdDox",
							"name": "a1",
							"email": "a1@example.com",
							"status": "PENDING",
							"createdAt": "2023-02-24T14:48:30Z",
							"additionalInfo": "af1"
						},
						{
							"id": "QWNjZXNzUmVxdWVzdDoy",
							"name": "a2",
							"email": "a2@example.com",
							"status": "APPROVED",
							"createdAt": "2023-02-24T14:48:30Z",
							"additionalInfo": ""
						},
						{
							"id": "QWNjZXNzUmVxdWVzdDoz",
							"name": "a3",
							"email": "a3@example.com",
							"status": "REJECTED",
							"createdAt": "2023-02-24T14:48:30Z",
							"additionalInfo": ""
						}
					],
					"totalCount": 3,
					"pageInfo": {
						"hasNextPage": false,
						"hasPreviousPage": false,
						"startCursor": "QWNjZXNzUmVxdWVzdDox",
						"endCursor": "QWNjZXNzUmVxdWVzdDoz"
					}
				}
			}`,
			Variables: map[string]any{
				"first": 10,
			},
		})
	})
}

func TestSetAccessRequestStatusMutation(t *testing.T) {
	const setAccessRequestStatusMutation = `
	mutation SetAccessRequestStatus($id: ID!, $status: AccessRequestStatus!) {
		setAccessRequestStatus(id: $id, status: $status) {
			alwaysNil
		}
	}`
	db := database.NewMockDB()
	db.WithTransactFunc.SetDefaultHook(func(ctx context.Context, f func(database.DB) error) error {
		return f(db)
	})

	userStore := database.NewMockUserStore()
	db.UsersFunc.SetDefaultReturn(userStore)
	mockDBClient := database.NewMockDBClient()
	db.ClientFunc.SetDefaultReturn(mockDBClient)

	t.Parallel()

	t.Run("non-admin user", func(t *testing.T) {
		userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Context:        ctx,
			Query:          setAccessRequestStatusMutation,
			ExpectedResult: `{"setAccessRequestStatus": null }`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"setAccessRequestStatus"},
					Message:       auth.ErrMustBeSiteAdmin.Error(),
					ResolverError: auth.ErrMustBeSiteAdmin,
				},
			},
			Variables: map[string]any{
				"id":     string(marshalAccessRequestID(1)),
				"status": string(types.AccessRequestStatusApproved),
			},
		})
		assert.Len(t, mockDBClient.History(&accessrequests.Update{}), 0)
	})

	t.Run("existing access request", func(t *testing.T) {
		createdAtTime, _ := time.Parse(time.RFC3339, "2023-02-24T14:48:30Z")
		mockAccessRequest := &types.AccessRequest{ID: 1, Email: "a1@example.com", Name: "a1", CreatedAt: createdAtTime, AdditionalInfo: "af1", Status: types.AccessRequestStatusPending}
		mockDBClient.Mock(&accessrequests.GetByID{}, mockAccessRequest, nil)
		mockDBClient.Mock(&accessrequests.Update{}, mockAccessRequest, nil)
		userID := int32(123)
		userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: userID, SiteAdmin: true}, nil)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Context:        ctx,
			Query:          setAccessRequestStatusMutation,
			ExpectedResult: `{"setAccessRequestStatus": { "alwaysNil": null } }`,
			Variables: map[string]any{
				"id":     string(marshalAccessRequestID(1)),
				"status": string(types.AccessRequestStatusApproved),
			},
		})
		assert.Len(t, mockDBClient.History(&accessrequests.Update{}), 1)
		assert.Equal(t, &accessrequests.Update{
			AccessRequest: &types.AccessRequest{
				ID:               mockAccessRequest.ID,
				DecisionByUserID: &userID,
				Status:           types.AccessRequestStatusApproved,
			},
		}, mockDBClient.History(&accessrequests.Update{})[0])
	})

	t.Run("non-existing access request", func(t *testing.T) {
		mockDBClient := database.NewMockDBClient()
		db.ClientFunc.SetDefaultReturn(mockDBClient)
		notFoundErr := &accessrequests.ErrNotFound{ID: 1}
		var notFound *types.AccessRequest
		mockDBClient.Mock(&accessrequests.GetByID{}, notFound, notFoundErr)

		userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Context:        ctx,
			Query:          setAccessRequestStatusMutation,
			ExpectedResult: `{"setAccessRequestStatus": null }`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"setAccessRequestStatus"},
					Message:       "access_request with ID 1 not found",
					ResolverError: notFoundErr,
				},
			},
			Variables: map[string]any{
				"id":     string(marshalAccessRequestID(1)),
				"status": string(types.AccessRequestStatusApproved),
			},
		})
		assert.Len(t, mockDBClient.History(&accessrequests.Update{}), 0)
	})
}
