package resolvers

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// TestResolver_InsightSeries tests that the InsightSeries GraphQL resolver works.
func TestResolver_InsightSeries(t *testing.T) {
	testSetup := func(t *testing.T) (context.Context, [][]graphqlbackend.InsightSeriesResolver) {
		// Setup the GraphQL resolver.
		ctx := actor.WithInternalActor(context.Background())
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond)
		logger := logtest.Scoped(t)
		clock := func() time.Time { return now }
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
		postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
		resolver := newWithClock(insightsDB, postgres, clock)
		insightStore := store.NewInsightStore(insightsDB)

		view := types.InsightView{
			Title:            "title1",
			Description:      "desc1",
			PresentationType: types.Line,
		}
		insightSeries := types.InsightSeries{
			SeriesID:            "1234567",
			Query:               "query1",
			CreatedAt:           now,
			OldestHistoricalAt:  now,
			LastRecordedAt:      now,
			NextRecordingAfter:  now,
			SampleIntervalUnit:  string(types.Month),
			SampleIntervalValue: 1,
		}
		var err error
		view, err = insightStore.CreateView(ctx, view, []store.InsightViewGrant{store.GlobalGrant()})
		require.NoError(t, err)
		insightSeries, err = insightStore.CreateSeries(ctx, insightSeries)
		require.NoError(t, err)
		insightStore.AttachSeriesToView(ctx, insightSeries, view, types.InsightViewSeriesMetadata{
			Label:  "",
			Stroke: "",
		})

		insightMetadataStore := store.NewMockInsightMetadataStore()

		resolver.insightMetadataStore = insightMetadataStore

		// Create the insightview connection resolver and query series.
		conn, err := resolver.InsightViews(ctx, &graphqlbackend.InsightViewQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		nodes, err := conn.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var series [][]graphqlbackend.InsightSeriesResolver
		for _, node := range nodes {
			s, _ := node.DataSeries(ctx)
			series = append(series, s)
		}
		return ctx, series
	}

	t.Run("Points", func(t *testing.T) {
		ctx, insights := testSetup(t)
		autogold.Want("insights length", int(1)).Equal(t, len(insights))

		autogold.Want("insights[0].length", int(1)).Equal(t, len(insights[0]))

		// Issue a query against the actual DB.
		points, err := insights[0][0].Points(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("insights[0][0].Points", []graphqlbackend.InsightsDataPointResolver{}).Equal(t, points)

	})
}

func Test_augmentPointsForRecordingTimes(t *testing.T) {
	stringify := func(points []store.SeriesPoint) []string {
		s := []string{}
		for _, point := range points {
			s = append(s, point.String())
		}
		// Sort for determinism.
		sort.Strings(s)
		return s
	}
	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	generateTimes := func(n int) []time.Time {
		times := []time.Time{}
		for i := 0; i < n; i++ {
			times = append(times, testTime.AddDate(0, 0, i))
		}
		return times
	}
	capture := func(s string) *string {
		return &s
	}
	makeRecordingTimes := func(times []time.Time) []types.RecordingTime {
		recordingTimes := make([]types.RecordingTime, len(times))
		for i, t := range times {
			recordingTimes[i] = types.RecordingTime{Timestamp: t}
		}
		return recordingTimes
	}

	testCases := []struct {
		points         []store.SeriesPoint
		recordingTimes []time.Time
		want           autogold.Value
	}{
		{
			nil,
			nil,
			autogold.Want("empty returns empty", []string{}),
		},
		{
			[]store.SeriesPoint{{"seriesID", testTime, 12, nil}},
			[]time.Time{},
			autogold.Want("empty recording times returns empty", []string{}),
		},
		{
			[]store.SeriesPoint{
				{"1", testTime, 1, nil},
			},
			generateTimes(2),
			autogold.Want("augment one data point", []string{
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Value: 1}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Value: 0}`,
			}),
		},
		{
			[]store.SeriesPoint{
				{"1", testTime, 1, capture("one")},
				{"1", testTime, 2, capture("two")},
				{"1", testTime, 3, capture("three")},
				{"1", testTime.AddDate(0, 0, 1), 1, capture("one")},
			},
			generateTimes(2),
			autogold.Want("augment capture data points", []string{
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Capture: "one", Value: 1}`,
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Capture: "three", Value: 3}`,
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Capture: "two", Value: 2}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Capture: "one", Value: 1}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Capture: "three", Value: 0}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Capture: "two", Value: 0}`,
			}),
		},
		{
			[]store.SeriesPoint{
				{"1", testTime, 11, nil},
				{"1", testTime.AddDate(0, 0, 1), 22, nil},
			},
			append([]time.Time{testTime.AddDate(0, 0, -1)}, generateTimes(2)...),
			autogold.Want("augment data point in the past", []string{
				`SeriesPoint{Time: "2019-12-31 00:00:00 +0000 UTC", Value: 0}`,
				`SeriesPoint{Time: "2020-01-01 00:00:00 +0000 UTC", Value: 11}`,
				`SeriesPoint{Time: "2020-01-02 00:00:00 +0000 UTC", Value: 22}`,
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := augmentPointsForRecordingTimes(tc.points, makeRecordingTimes(tc.recordingTimes))
			tc.want.Equal(t, stringify(got))
		})
	}
}
