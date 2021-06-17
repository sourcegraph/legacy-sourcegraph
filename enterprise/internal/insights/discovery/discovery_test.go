package discovery

import (
	"context"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

var settingsExample = &api.Settings{ID: 1, Contents: `{
	"insights": [
		{
		  "title": "fmt usage",
		  "description": "fmt.Errorf/fmt.Printf usage",
		  "series": [
			{
			  "label": "fmt.Errorf",
			  "search": "errorf",
			},
			{
			  "label": "printf",
			  "search": "fmt.Printf",
			}
		  ]
		},
		{
			"title": "gitserver usage",
			"description": "gitserver exec & close usage",
			"series": [
			  {
				"label": "exec",
				"search": "gitserver.Exec",
			  },
			  {
				"label": "close",
				"search": "gitserver.Close",
			  }
			]
		  }
		]
	}
`}

func TestDiscover(t *testing.T) {
	settingStore := NewMockSettingStore()
	settingStore.GetLatestFunc.SetDefaultHook(func(ctx context.Context, subject api.SettingsSubject) (*api.Settings, error) {
		if !subject.Site { // TODO: future: site is an extremely poor name for "global settings", we should change this.
			t.Fatal("expected only to request settings from global user settings")
		}
		return settingsExample, nil
	})
	ctx := context.Background()
	insights, err := Discover(ctx, settingStore)
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("insights", []*schema.Insight{
		{
			Description: "fmt.Errorf/fmt.Printf usage",
			Series: []*schema.InsightSeries{
				{
					Label:  "fmt.Errorf",
					Search: "errorf",
				},
				{
					Label:  "printf",
					Search: "fmt.Printf",
				},
			},
			Title: "fmt usage",
		},
		{
			Description: "gitserver exec & close usage",
			Series: []*schema.InsightSeries{
				{
					Label:  "exec",
					Search: "gitserver.Exec",
				},
				{
					Label:  "close",
					Search: "gitserver.Close",
				},
			},
			Title: "gitserver usage",
		},
	}).Equal(t, insights)
}
