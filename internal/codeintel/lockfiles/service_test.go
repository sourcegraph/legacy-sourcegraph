package lockfiles

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestListDependencies(t *testing.T) {
	lsFiles := func(context.Context, authz.SubRepoPermissionChecker, api.RepoName, api.CommitID, ...string) ([]string, error) {
		return []string{"client/package-lock.json", "package-lock.json"}, nil
	}

	archive := func(c context.Context, repo api.RepoName, ao gitserver.ArchiveOptions) (io.ReadCloser, error) {
		var b bytes.Buffer
		zw := zip.NewWriter(&b)
		defer zw.Close()

		for file, data := range map[string]string{
			"client/package-lock.json": `{"dependencies": { "@octokit/request": {"version": "5.6.2"} }}`,
			"package-lock.json":        `{"dependencies": { "nan": {"version": "2.15.0"} }}`,
		} {
			w, err := zw.Create(file)
			if err != nil {
				t.Fatal(err)
			}

			_, err = w.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		}

		return io.NopCloser(&b), nil
	}

	s := TestService(
		authz.DefaultSubRepoPermsChecker,
		lsFiles,
		archive,
	)

	ctx := context.Background()
	got, err := s.ListDependencies(ctx, "foo", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	want := []reposource.PackageDependency{
		npmDependency(t, "@octokit/request@5.6.2"),
		npmDependency(t, "nan@2.15.0"),
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i].PackageManagerSyntax() < got[j].PackageManagerSyntax()
	})

	comparer := cmp.Comparer(func(a, b reposource.PackageDependency) bool {
		return a.PackageManagerSyntax() == b.PackageManagerSyntax()
	})

	if diff := cmp.Diff(want, got, comparer); diff != "" {
		t.Fatalf("dependency mismatch (-want +got):\n%s", diff)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		file string
		data string
		want []reposource.PackageDependency
	}{
		{
			file: "package-lock.json",
			data: `{"dependencies": {
        "nan": {"version": "2.15.0"},
        "@octokit/request": {"version": "5.6.2"}
      }}`,
			want: []reposource.PackageDependency{
				npmDependency(t, "@octokit/request@5.6.2"),
				npmDependency(t, "nan@2.15.0"),
			},
		},
		{
			file: "yarn.lock",
			data: `
# THIS IS AN AUTOGENERATED FILE. DO NOT EDIT THIS FILE DIRECTLY.
# yarn lockfile v1

"@actions/core@^1.2.4":
  version "1.2.6"
  resolved "https://registry.npmjs.org/@actions/core/-/core-1.2.6.tgz#a78d49f41a4def18e88ce47c2cac615d5694bf09"
  integrity sha512-ZQYitnqiyBc3D+k7LsgSBmMDVkOVidaagDG7j3fOym77jNunWRuYx7VSHa9GNfFZh+zh61xsCjRj4JxMZlDqTA==

"@actions/github@^4.0.0":
  version "4.0.0"
  resolved "https://registry.npmjs.org/@actions/github/-/github-4.0.0.tgz#d520483151a2bf5d2dc9cd0f20f9ac3a2e458816"
  integrity sha512-Ej/Y2E+VV6sR9X7pWL5F3VgEWrABaT292DRqRU6R4hnQjPtC/zD3nagxVdXWiRQvYDh8kHXo7IDmG42eJ/dOMA==
  dependencies:
    "@actions/http-client" "^1.0.8"
    "@octokit/core" "^3.0.0"
    "@octokit/plugin-paginate-rest" "^2.2.3"
    "@octokit/plugin-rest-endpoint-methods" "^4.0.0"
`,
			want: []reposource.PackageDependency{
				npmDependency(t, "@actions/core@1.2.6"),
				npmDependency(t, "@actions/github@4.0.0"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			got, err := parse(test.file, strings.NewReader(test.data))
			if err != nil {
				t.Fatal(err)
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].PackageManagerSyntax() < got[j].PackageManagerSyntax()
			})

			comparer := cmp.Comparer(func(a, b reposource.PackageDependency) bool {
				return a.PackageManagerSyntax() == b.PackageManagerSyntax()
			})

			if diff := cmp.Diff(test.want, got, comparer); diff != "" {
				t.Fatalf("dependency mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func npmDependency(t testing.TB, dep string) *reposource.NPMDependency {
	t.Helper()

	d, err := reposource.ParseNPMDependency(dep)
	if err != nil {
		t.Fatal(err)
	}

	return d
}
