package main

import (
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"testing"
)

func TestModeAvailability(t *testing.T) {
	t.Parallel()

	t.Run("invalid query returns unavailable", func(t *testing.T) {
		availabilities, err := client.ModeAvailability("fork:insights test", "literal")
		if err != nil {
			t.Fatal(err)
		}
		for _, response := range availabilities {
			if response.Available == true {
				t.Errorf("expected mode %v to be unavailable", response.Mode)
			}
			if response.ReasonUnavailable == nil {
				t.Errorf("expected to receive an unavailable reason, got nil")
			}
		}
	})

	t.Run("returns repo path capture group", func(t *testing.T) {
		query := `(\w)\s\*testing.T`
		availabilities, err := client.ModeAvailability(query, "regexp")
		if err != nil {
			t.Fatal(err)
		}
		for mode, response := range availabilities {
			if mode == "REPO" || mode == "PATH" || mode == "CAPTURE_GROUP" {
				if response.Available != true {
					t.Errorf("expected mode %v to be available for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable != nil {
					t.Errorf("expected to be available, got %q", *response.ReasonUnavailable)
				}
			} else {
				if response.Available == true {
					t.Errorf("expected mode %v to be unavailable for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable == nil {
					t.Errorf("expected to receive an unavailable reason, got nil")
				}
			}
		}
	})

	t.Run("returns repo author", func(t *testing.T) {
		query := "type:commit insights"
		availabilities, err := client.ModeAvailability(query, "literal")
		if err != nil {
			t.Fatal(err)
		}
		for mode, response := range availabilities {
			if mode == "REPO" || mode == "AUTHOR" {
				if response.Available != true {
					t.Errorf("expected mode %v to be available for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable != nil {
					t.Errorf("expected to be available, got %q", *response.ReasonUnavailable)
				}
			} else {
				if response.Available == true {
					t.Errorf("expected mode %v to be unavailable for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable == nil {
					t.Errorf("expected to receive an unavailable reason, got nil")
				}
			}
		}
	})
}

func TestAggregations(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	_, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-github-search",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/mux",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = client.WaitForReposToBeCloned(
		"sgtest/mux",
	)
	if err != nil {
		t.Fatal(err)
	}

	err = client.WaitForReposToBeIndexed(
		"sgtest/mux",
	)
	if err != nil {
		t.Fatal(err)
	}
}
