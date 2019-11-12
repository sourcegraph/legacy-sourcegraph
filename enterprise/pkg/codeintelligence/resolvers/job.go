package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintelligence/lsif"
)

//
// Node Resolver

type lsifJobResolver struct {
	lsifJob *types.LSIFJob
}

var _ graphqlbackend.LSIFJobResolver = &lsifJobResolver{}

func (r *lsifJobResolver) ID() graphql.ID {
	return marshalLSIFJobGQLID(r.lsifJob.ID)
}

func (r *lsifJobResolver) Name() string {
	return r.lsifJob.Name
}

func (r *lsifJobResolver) Args() graphqlbackend.JSONValue {
	return graphqlbackend.JSONValue{r.lsifJob.Args}
}

func (r *lsifJobResolver) State() string {
	return strings.ToUpper(r.lsifJob.State)
}

func (r *lsifJobResolver) Progress() float64 {
	return r.lsifJob.Progress
}

func (r *lsifJobResolver) FailedReason() *string {
	return r.lsifJob.FailedReason
}

func (r *lsifJobResolver) Stacktrace() *[]string {
	return r.lsifJob.Stacktrace
}

func (r *lsifJobResolver) Timestamp() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.lsifJob.Timestamp}
}

func (r *lsifJobResolver) ProcessedOn() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(r.lsifJob.ProcessedOn)
}

func (r *lsifJobResolver) FinishedOn() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(r.lsifJob.FinishedOn)
}

//
// Connection Resolver

type lsifJobConnectionResolver struct {
	opt LSIFJobsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	jobs       []*types.LSIFJob
	totalCount *int
	nextURL    string
	err        error
}

var _ graphqlbackend.LSIFJobConnectionResolver = &lsifJobConnectionResolver{}

func (r *lsifJobConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFJobResolver, error) {
	jobs, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LSIFJobResolver
	for _, lsifJob := range jobs {
		l = append(l, &lsifJobResolver{
			lsifJob: lsifJob,
		})
	}
	return l, nil
}

func (r *lsifJobConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	_, count, _, err := r.compute(ctx)
	if count == nil || err != nil {
		return nil, err
	}

	c := int32(*count)
	return &c, nil
}

func (r *lsifJobConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *lsifJobConnectionResolver) compute(ctx context.Context) ([]*types.LSIFJob, *int, string, error) {
	r.once.Do(func() {
		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/jobs/%s", strings.ToLower(r.opt.State))
		} else {
			// subsequent page of results
			path = *r.opt.NextURL
		}

		query := url.Values{}
		if r.opt.Query != nil {
			query.Set("query", *r.opt.Query)
		}
		if r.opt.Limit != nil {
			query.Set("limit", strconv.FormatInt(int64(*r.opt.Limit), 10))
		}

		resp, err := lsif.BuildAndTraceRequest(ctx, path, query)
		if err != nil {
			r.err = err
			return
		}

		payload := struct {
			Jobs       []*types.LSIFJob `json:"jobs"`
			TotalCount *int             `json:"totalCount"`
		}{
			Jobs: []*types.LSIFJob{},
		}

		if err := lsif.UnmarshalPayload(resp, &payload); err != nil {
			r.err = err
			return
		}

		r.jobs = payload.Jobs
		r.totalCount = payload.TotalCount
		r.nextURL = lsif.ExtractNextURL(resp)
	})

	return r.jobs, r.totalCount, r.nextURL, r.err
}

//
// Stats Resolver

type lsifJobStatsResolver struct {
	stats *types.LSIFJobStats
}

var _ graphqlbackend.LSIFJobStatsResolver = &lsifJobStatsResolver{}

func (r *lsifJobStatsResolver) ID() graphql.ID {
	return marshalLSIFJobStatsGQLID(lsifJobStatsGQLID)
}

func (r *lsifJobStatsResolver) ProcessingCount() int32 { return r.stats.ProcessingCount }
func (r *lsifJobStatsResolver) ErroredCount() int32    { return r.stats.ErroredCount }
func (r *lsifJobStatsResolver) CompletedCount() int32  { return r.stats.CompletedCount }
func (r *lsifJobStatsResolver) QueuedCount() int32     { return r.stats.QueuedCount }
func (r *lsifJobStatsResolver) ScheduledCount() int32  { return r.stats.ScheduledCount }

//
// ID Serialization

func marshalLSIFJobGQLID(lsifJobID string) graphql.ID {
	return relay.MarshalID("LSIFJob", lsifJobID)
}

func unmarshalLSIFJobGQLID(id graphql.ID) (lsifJobID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobID)
	return
}

func marshalLSIFJobStatsGQLID(lsifJobStatsID string) graphql.ID {
	return relay.MarshalID("LSIFJobStats", lsifJobStatsID)
}

func unmarshalLSIFJobStatsGQLID(id graphql.ID) (lsifJobStatsID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobStatsID)
	return
}
