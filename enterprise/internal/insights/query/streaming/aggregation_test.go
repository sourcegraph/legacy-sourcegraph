package streaming

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestAddAggregate(t *testing.T) {
	testCases := []struct {
		name  string
		have  aggregated
		value string
		count int32
		want  aggregated
	}{
		{
			name: "adds up other count",
			have: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
			},
			value: "B",
			count: 9,
			want: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
				OtherCount:       OtherCount{ResultCount: 9, GroupCount: 1},
			},
		},
		{
			name: "adds new result",
			have: aggregated{
				resultBufferSize: 2,
				Results:          map[string]int32{"A": 24},
				smallestResult:   &Aggregate{"A", 24},
			},
			value: "B",
			count: 32,
			want: aggregated{
				resultBufferSize: 2,
				Results:          map[string]int32{"A": 24, "B": 32},
				smallestResult:   &Aggregate{"A", 24},
			},
		},
		{
			name: "updates existing results",
			have: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"C": 5},
				smallestResult:   &Aggregate{"C", 5},
			},
			value: "C",
			count: 11,
			want: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"C": 16},
				smallestResult:   &Aggregate{"C", 16},
			},
		},
		{
			name: "ejects smallest result",
			have: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"C": 5},
				smallestResult:   &Aggregate{"C", 5},
			},
			value: "A",
			count: 15,
			want: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 15},
				smallestResult:   &Aggregate{"A", 15},
				OtherCount:       OtherCount{ResultCount: 5, GroupCount: 1},
			},
		},
		{
			name: "adds up other group count",
			have: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
				OtherCount:       OtherCount{ResultCount: 9, GroupCount: 1},
			},
			value: "B",
			count: 9,
			want: aggregated{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
				OtherCount:       OtherCount{ResultCount: 18, GroupCount: 2},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.have.Add(tc.value, tc.count)
			autogold.Want(tc.name, tc.want).Equal(t, tc.have)
		})
	}
}

func TestFindSmallestAggregate(t *testing.T) {
	testCases := []struct {
		name string
		have aggregated
		want *Aggregate
	}{
		{
			name: "returns nil for empty results",
			want: nil,
		},
		{
			name: "one result is smallest",
			have: aggregated{
				Results: map[string]int32{"myresult": 20},
			},
			want: &Aggregate{"myresult", 20},
		},
		{
			name: "finds smallest result by count",
			have: aggregated{
				Results: map[string]int32{"high": 20, "low": 5, "mid": 10},
			},
			want: &Aggregate{"low", 5},
		},
		{
			name: "finds smallest result by label",
			have: aggregated{
				Results: map[string]int32{"outsider": 5, "abc/1": 5, "abcd": 5, "abc/2": 5},
			},
			want: &Aggregate{"abc/1", 5},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.have.findSmallestAggregate()
			autogold.Want("smallest result should match", tc.want).Equal(t, got)
		})
	}
}

func TestSortAggregate(t *testing.T) {
	a := aggregated{
		Results:          make(map[string]int32),
		resultBufferSize: 5,
	}

	// Add 5 distinct elements. Update 1 existing.
	a.Add("sg/1", 5)
	a.Add("sg/2", 10)
	a.Add("sg/3", 8)
	a.Add("sg/1", 3)
	a.Add("sg/4", 22)
	a.Add("sg/5", 60)

	// Add two more elements.
	a.Add("sg/too-much", 12)
	a.Add("sg/lost", 1)

	// Update another one.
	a.Add("sg/2", 5)

	autogold.Want("other result count should be 13", int32(13)).Equal(t, a.OtherCount.ResultCount)
	autogold.Want("other group count should be 2", int32(2)).Equal(t, a.OtherCount.GroupCount)

	want := []*Aggregate{
		{"sg/5", 60},
		{"sg/4", 22},
		{"sg/2", 15},
		{"sg/1", 8},
		{"sg/3", 8},
	}
	autogold.Want("SortAggregate should return DESC sorted list", want).Equal(t, a.SortAggregate())
}
