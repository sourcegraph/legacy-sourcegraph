package repos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketServerSource_MakeRepo(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketServerConnection{
		"simple": {
			Url:   "bitbucket.example.com",
			Token: "secret",
		},
		"ssh": {
			Url:                         "https://bitbucket.example.com",
			Token:                       "secret",
			InitialRepositoryEnablement: true,
			GitURLType:                  "ssh",
		},
		"path-pattern": {
			Url:                   "https://bitbucket.example.com",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
		"username": {
			Url:                   "https://bitbucket.example.com",
			Username:              "foo",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketServer}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketServerSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*types.Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r, false))
			}

			path := filepath.Join("testdata", "bitbucketserver-repos-"+name+".golden")
			testutil.AssertGolden(t, path, update(name), got)
		})
	}
}

func TestBitbucketServerSource_Exclude(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketServerConnection{
		"none": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
		},
		"name": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Name: "SG/python-langserver-fork",
			}, {
				Name: "~KEEGAN/rgp",
			}},
		},
		"id": {
			Url:     "https://bitbucket.example.com",
			Token:   "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{Id: 4}},
		},
		"pattern": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Pattern: "SG/python.*",
			}, {
				Pattern: "~KEEGAN/.*",
			}},
		},
		"both": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			// We match on the bitbucket server repo name, not the repository path pattern.
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Id: 1,
			}, {
				Name: "~KEEGAN/rgp",
			}, {
				Pattern: ".*-fork",
			}},
		},
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketServer}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketServerSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			type output struct {
				Include []string
				Exclude []string
			}
			var got output
			for _, r := range repos {
				name := r.Slug
				if r.Project != nil {
					name = r.Project.Key + "/" + name
				}
				if s.excludes(r) {
					got.Exclude = append(got.Exclude, name)
				} else {
					got.Include = append(got.Include, name)
				}
			}

			path := filepath.Join("testdata", "bitbucketserver-repos-exclude-"+name+".golden")
			testutil.AssertGolden(t, path, update(name), got)
		})
	}
}

func TestBitbucketServerSource_WithAuthenticator(t *testing.T) {
	svc := &types.ExternalService{
		Kind: extsvc.KindBitbucketServer,
		Config: marshalJSON(t, &schema.BitbucketServerConnection{
			Url:   "https://bitbucket.sgdev.org",
			Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
		}),
	}

	bbsSrc, err := NewBitbucketServerSource(svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("supported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"BasicAuth":           &auth.BasicAuth{},
			"OAuthBearerToken":    &auth.OAuthBearerToken{},
			"SudoableOAuthClient": &bitbucketserver.SudoableOAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticator(tc)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				if gs, ok := src.(*BitbucketServerSource); !ok {
					t.Error("cannot coerce Source into bbsSource")
				} else if gs == nil {
					t.Error("unexpected nil Source")
				}
			})
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"nil":         nil,
			"OAuthClient": &auth.OAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticator(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HasType(err, UnsupportedAuthenticatorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestBitbucketServerSource_ListRepos(t *testing.T) {
	TestBitbucketServerSource_ListByReposOnly(t)
	TestBitbucketServerSource_ListByRepositoryQueryDefault(t)
	TestBitbucketServerSource_ListByRepositoryQueryAll(t)
	TestBitbucketServerSource_ListByRepositoryQueryNone(t)
	TestBitbucketServerSource_ListByProjectKey(t)
}

func TestBitbucketServerSource_ListByReposOnly(t *testing.T) {
	repos := GetReposFromTestdata(t)

	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/1.0/projects/", func(w http.ResponseWriter, r *http.Request) {
		pathArr := strings.Split(r.URL.Path, "/")
		projectKey := pathArr[5]
		repoSlug := pathArr[7]

		for _, repo := range repos {
			if repo.Project.Key == projectKey && repo.Slug == repoSlug {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(repo)
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cases, svc := GetConfig(t, server)
	for _, config := range cases {
		s, err := newBitbucketServerSource(&svc, config, nil)
		if err != nil {
			t.Fatal(err)
		}

		s.config.Repos = []string{
			"SG/go-langserver",
			"SG/python-langserver",
			"SG/python-langserver-fork",
			"~KEEGAN/rgp",
			"~KEEGAN/rgp-unavailable",
		}

		ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunction()

		results := make(chan SourceResult, 10)
		defer close(results)

		s.ListRepos(ctxWithTimeout, results)
		VerifyData(t, ctxWithTimeout, 4, results)
	}
}

func TestBitbucketServerSource_ListByRepositoryQueryDefault(t *testing.T) {
	repos := GetReposFromTestdata(t)

	type Results struct {
		*bitbucketserver.PageToken
		Values any `json:"values"`
	}

	pageToken := bitbucketserver.PageToken{
		Size:          1,
		Limit:         1000,
		IsLastPage:    true,
		Start:         1,
		NextPageStart: 1,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/1.0/repos", func(w http.ResponseWriter, r *http.Request) {
		projectName := r.URL.Query().Get("projectName")

		if projectName == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Results{
				PageToken: &pageToken,
				Values:    repos,
			})
		} else {
			for _, repo := range repos {
				if projectName == repo.Name {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(Results{
						PageToken: &pageToken,
						Values:    []bitbucketserver.Repo{repo},
					})
				}
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cases, svc := GetConfig(t, server)
	for _, config := range cases {
		s, err := newBitbucketServerSource(&svc, config, nil)
		if err != nil {
			t.Fatal(err)
		}

		s.config.RepositoryQuery = []string{
			"?projectName=go-langserver",
			"?projectName=python-langserver",
			"?projectName=python-langserver-fork",
			"?projectName=rgp",
			"?projectName=rgp-unavailable",
		}

		ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunction()

		results := make(chan SourceResult, 20)
		defer close(results)

		s.ListRepos(ctxWithTimeout, results)
		VerifyData(t, ctxWithTimeout, 4, results)
	}
}

func TestBitbucketServerSource_ListByRepositoryQueryAll(t *testing.T) {
	repos := GetReposFromTestdata(t)

	type Results struct {
		*bitbucketserver.PageToken
		Values any `json:"values"`
	}

	pageToken := bitbucketserver.PageToken{
		Size:          1,
		Limit:         1000,
		IsLastPage:    true,
		Start:         1,
		NextPageStart: 1,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/1.0/repos", func(w http.ResponseWriter, r *http.Request) {
		projectName := r.URL.Query().Get("projectName")

		if projectName == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Results{
				PageToken: &pageToken,
				Values:    repos,
			})
		} else {
			for _, repo := range repos {
				if projectName == repo.Name {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(Results{
						PageToken: &pageToken,
						Values:    []bitbucketserver.Repo{repo},
					})
				}
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cases, svc := GetConfig(t, server)
	for _, config := range cases {
		s, err := newBitbucketServerSource(&svc, config, nil)
		if err != nil {
			t.Fatal(err)
		}

		s.config.RepositoryQuery = []string{
			"all",
		}

		ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunction()

		results := make(chan SourceResult, 20)
		defer close(results)

		s.ListRepos(ctxWithTimeout, results)
		VerifyData(t, ctxWithTimeout, 4, results)
	}
}

func TestBitbucketServerSource_ListByRepositoryQueryNone(t *testing.T) {
	repos := GetReposFromTestdata(t)

	type Results struct {
		*bitbucketserver.PageToken
		Values any `json:"values"`
	}

	pageToken := bitbucketserver.PageToken{
		Size:          1,
		Limit:         1000,
		IsLastPage:    true,
		Start:         1,
		NextPageStart: 1,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/1.0/repos", func(w http.ResponseWriter, r *http.Request) {
		projectName := r.URL.Query().Get("projectName")

		if projectName == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Results{
				PageToken: &pageToken,
				Values:    repos,
			})
		} else {
			for _, repo := range repos {
				if projectName == repo.Name {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(Results{
						PageToken: &pageToken,
						Values:    []bitbucketserver.Repo{repo},
					})
				}
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cases, svc := GetConfig(t, server)
	for _, config := range cases {
		s, err := newBitbucketServerSource(&svc, config, nil)
		if err != nil {
			t.Fatal(err)
		}

		s.config.RepositoryQuery = []string{
			"none",
		}

		ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunction()

		results := make(chan SourceResult, 20)
		defer close(results)

		s.ListRepos(ctxWithTimeout, results)
		VerifyData(t, ctxWithTimeout, 0, results)
	}
}

func TestBitbucketServerSource_ListByProjectKey(t *testing.T) {
	repos := GetReposFromTestdata(t)

	type Results struct {
		*bitbucketserver.PageToken
		Values any `json:"values"`
	}

	pageToken := bitbucketserver.PageToken{
		Size:          1,
		Limit:         1000,
		IsLastPage:    true,
		Start:         1,
		NextPageStart: 1,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/1.0/projects/", func(w http.ResponseWriter, r *http.Request) {
		pathArr := strings.Split(r.URL.Path, "/")
		projectKey := pathArr[5]
		values := make([]bitbucketserver.Repo, 0)

		for _, repo := range repos {
			if repo.Project.Key == projectKey {
				values = append(values, repo)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Results{
			PageToken: &pageToken,
			Values:    values,
		})

	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cases, svc := GetConfig(t, server)
	for _, config := range cases {
		s, err := newBitbucketServerSource(&svc, config, nil)
		if err != nil {
			t.Fatal(err)
		}

		s.config.ProjectKeys = []string{
			"SG",
			"~KEEGAN",
		}

		ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunction()

		results := make(chan SourceResult, 20)
		defer close(results)

		s.ListRepos(ctxWithTimeout, results)
		VerifyData(t, ctxWithTimeout, 4, results)
	}
}

func GetReposFromTestdata(t *testing.T) []bitbucketserver.Repo {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}

	var repos []bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	return repos
}

func GetConfig(t *testing.T, server *httptest.Server) (map[string]*schema.BitbucketServerConnection, types.ExternalService) {
	cases := map[string]*schema.BitbucketServerConnection{
		"simple": {
			Url:   server.URL,
			Token: "secret",
		},
		"ssh": {
			Url:                         server.URL,
			Token:                       "secret",
			InitialRepositoryEnablement: true,
			GitURLType:                  "ssh",
		},
		"path-pattern": {
			Url:                   server.URL,
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
		"username": {
			Url:                   server.URL,
			Username:              "foo",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketServer}

	return cases, svc
}

func VerifyData(t *testing.T, ctx context.Context, numExpectedResults int, results chan SourceResult) {
	numTotalResults := len(results)
	numReceivedFromResults := 0

	if numTotalResults != numExpectedResults {
		t.Fatal(errors.New("wrong number of results"))
	}

	repoNameMap := map[string]struct{}{
		"SG/go-langserver":          {},
		"SG/python-langserver":      {},
		"SG/python-langserver-fork": {},
		"~KEEGAN/rgp":               {},
		"~KEEGAN/rgp-unavailable":   {},
	}

	for {
		select {
		case res := <-results:
			repoNameArr := strings.Split(string(res.Repo.Name), "/")
			repoName := repoNameArr[1] + "/" + repoNameArr[2]
			if _, ok := repoNameMap[repoName]; ok {
				numReceivedFromResults++
			} else {
				t.Fatal(errors.New("wrong repo returned"))
			}
		case <-ctx.Done():
			t.Fatal(errors.New("timeout!"))
		default:
			if numReceivedFromResults == numExpectedResults {
				return
			}
		}
	}
}
