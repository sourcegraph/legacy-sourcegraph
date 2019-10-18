package git_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestRepository_RawLogDiffSearch(t *testing.T) {
	t.Parallel()

	repo := MakeGitRepository(t,
		"echo root > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m root --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag mytag HEAD",

		"git checkout -b branch1",
		"echo branch1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m branch1 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",

		"git checkout -b branch2",
		"echo branch2 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m branch2 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
	)
	tests := []struct {
		name string
		opt  git.RawLogDiffSearchOptions
		want []*git.LogCommitSearchResult
	}{{
		name: "query",
		opt: git.RawLogDiffSearchOptions{
			Query: git.TextSearchOptions{Pattern: "root"},
			Diff:  true,
		},
		want: []*git.LogCommitSearchResult{{
			Commit: git.Commit{
				ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "branch1",
				Parents:   []api.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
			},
			Refs:       []string{"refs/heads/branch1"},
			SourceRefs: []string{"refs/heads/branch2"},
			Diff:       &git.Diff{Raw: "diff --git a/f b/f\nindex d8649da..1193ff4 100644\n--- a/f\n+++ b/f\n@@ -1,1 +1,1 @@\n-root\n+branch1\n"},
		}, {
			Commit: git.Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/heads/branch2"},
			Diff:       &git.Diff{Raw: "diff --git a/f b/f\nnew file mode 100644\nindex 0000000..d8649da\n--- /dev/null\n+++ b/f\n@@ -0,0 +1,1 @@\n+root\n"},
		}},
	}, {
		name: "empty-query",
		opt: git.RawLogDiffSearchOptions{
			Query: git.TextSearchOptions{Pattern: ""},
			Args:  []string{"--grep=branch1|root", "--extended-regexp"},
		},
		want: []*git.LogCommitSearchResult{{
			Commit: git.Commit{
				ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "branch1",
				Parents:   []api.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
			},
			Refs:       []string{"refs/heads/branch1"},
			SourceRefs: []string{"refs/heads/branch2"},
		}, {
			Commit: git.Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/heads/branch2"},
		}},
	}, {
		name: "path",
		opt: git.RawLogDiffSearchOptions{
			Paths: git.PathOptions{
				IncludePatterns: []string{"g"},
				ExcludePattern:  "f",
				IsRegExp:        true,
			},
		},
		want: nil, // empty
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, complete, err := git.RawLogDiffSearch(ctx, repo, test.opt)
			if err != nil {
				t.Fatal(err)
			}
			if !complete {
				t.Fatal("!complete")
			}
			for _, r := range results {
				r.DiffHighlights = nil // Highlights is tested separately
			}
			if !cmp.Equal(test.want, results) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(results, test.want))
			}
		})
	}
}

func TestRepository_RawLogDiffSearch_emptyCommit(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m empty --allow-empty --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m empty --allow-empty --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo gitserver.Repo
		want map[*git.RawLogDiffSearchOptions][]*git.LogCommitSearchResult
	}{
		"git cmd": {
			repo: MakeGitRepository(t, gitCommands...),
			want: map[*git.RawLogDiffSearchOptions][]*git.LogCommitSearchResult{
				{
					Paths: git.PathOptions{IncludePatterns: []string{"/xyz.txt"}, IsRegExp: true},
				}: nil, // want no matches
			},
		},
	}

	for label, test := range tests {
		for opt, want := range test.want {
			results, complete, err := git.RawLogDiffSearch(ctx, test.repo, *opt)
			if err != nil {
				t.Errorf("%s: %+v: %s", label, *opt, err)
				continue
			}
			if !complete {
				t.Errorf("%s: !complete", label)
			}
			for _, r := range results {
				r.DiffHighlights = nil // Highlights is tested separately
			}
			if !reflect.DeepEqual(results, want) {
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, AsJSON(results), AsJSON(want))
			}
		}
	}
}
