package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	got := (&GitTreeEntryResolver{
		commit: &GitCommitResolver{
			repo: &RepositoryResolver{
				repo: &types.Repo{Name: "my/repo"},
			},
		},
		stat: CreateFileInfo("a/b", true),
	}).RawZipArchiveURL()
	want := "http://example.com/my/repo/-/raw/a/b?format=zip"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGitTreeEntry_Content(t *testing.T) {
	wantPath := "foobar.md"
	wantContent := "foobar"

	git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
		if name != testPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", testPath, name)
		}
		return []byte(testContent), nil
	}
	t.Cleanup(func() { git.Mocks.ReadFile = nil })

	gitTree := &GitTreeEntryResolver{
		commit: &GitCommitResolver{
			repo: &RepositoryResolver{
				repo: &types.Repo{Name: "my/repo"},
			},
		},
		stat: CreateFileInfo(testPath, true),
	}

	newFileContent, err := gitTree.Content(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(newFileContent, testContent); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}
}
