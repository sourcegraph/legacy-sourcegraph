package syncer

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

func TestNextSync(t *testing.T) {
	t.Parallel()

	clock := func() time.Time { return time.Date(2020, 01, 01, 01, 01, 01, 01, time.UTC) }
	tests := []struct {
		name string
		h    *campaigns.ChangesetSyncData
		want time.Time
	}{
		{
			name: "No time passed",
			h: &campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock(),
			},
			want: clock().Add(minSyncDelay),
		},
		{
			name: "Linear backoff",
			h: &campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-1 * time.Hour),
			},
			want: clock().Add(1 * time.Hour),
		},
		{
			name: "Use max of ExternalUpdateAt and LatestEvent",
			h: &campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-2 * time.Hour),
				LatestEvent:       clock().Add(-1 * time.Hour),
			},
			want: clock().Add(1 * time.Hour),
		},
		{
			name: "Diff max is capped",
			h: &campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-2 * maxSyncDelay),
			},
			want: clock().Add(maxSyncDelay),
		},
		{
			name: "Diff min is capped",
			h: &campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-1 * minSyncDelay / 2),
			},
			want: clock().Add(minSyncDelay),
		},
		{
			name: "Event arrives after sync",
			h: &campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-1 * maxSyncDelay / 2),
				LatestEvent:       clock().Add(10 * time.Minute),
			},
			want: clock().Add(10 * time.Minute).Add(minSyncDelay),
		},
		{
			name: "Never synced",
			h:    &campaigns.ChangesetSyncData{},
			want: clock(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextSync(clock, tt.h)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
