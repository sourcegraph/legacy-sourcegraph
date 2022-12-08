package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/inventory"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGitTreeEntry_RawZipArchiveURL(t *testing.T) {
	db := database.NewMockDB()
	gitserverClient := gitserver.NewMockClient()
	got := NewGitTreeEntryResolver(db, gitserverClient,
		&GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		CreateFileInfo("a/b", true)).
		RawZipArchiveURL()
	want := "http://example.com/my/repo/-/raw/a/b?format=zip"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGitTreeEntry_Content(t *testing.T) {
	wantPath := "foobar.md"
	wantContent := "foobar"

	db := database.NewMockDB()
	gitserverClient := gitserver.NewMockClient()

	gitserverClient.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, _ api.CommitID, name string, _ authz.SubRepoPermissionChecker) ([]byte, error) {
		if name != wantPath {
			t.Fatalf("wrong name in ReadFile call. want=%q, have=%q", wantPath, name)
		}
		return []byte(wantContent), nil
	})

	gitTree := NewGitTreeEntryResolver(db, gitserverClient,
		&GitCommitResolver{
			repoResolver: NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}),
		},
		CreateFileInfo(wantPath, true))

	newFileContent, err := gitTree.Content(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(newFileContent, wantContent); diff != "" {
		t.Fatalf("wrong newFileContent: %s", diff)
	}

	newByteSize, err := gitTree.ByteSize(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if have, want := newByteSize, int32(len([]byte(wantContent))); have != want {
		t.Fatalf("wrong file size, want=%d have=%d", want, have)
	}
}

func TestGitTreeEntry_Stats(t *testing.T) {
	wantPath := "foobar.js"
	wantInventory := &inventory.Inventory{
		Languages: []inventory.Lang{{
			Name:       "JavaScript",
			TotalBytes: 555,
			TotalLines: 10,
		}},
	}
	wantLangStats := []inventory.Lang{{
		Name:       "JavaScript",
		TotalBytes: 555,
		TotalLines: 10,
	}}

	backend.Mocks.Repos.GetInventory = func(ctx context.Context, repo *types.Repo, commitID api.CommitID) (*inventory.Inventory, error) {
		if repo.Name != "my/repo" {
			t.Errorf("expected repo name %s, got %s", "my/repo", repo.Name)
		}
		if commitID != "aaaa" {
			t.Errorf("expected commit ID %s, got %s", "aaaa", commitID)
		}
		return wantInventory, nil
	}
	backend.Mocks.Repos.Get = func(v0 context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{Name: "my/repo"}, nil
	}
	rs := database.NewMockRepoStore()
	rs.GetFunc.SetDefaultReturn(&types.Repo{Name: "my/repo"}, nil)
	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(rs)
	gitserverClient := gitserver.NewMockClient()

	gitTree := NewGitTreeEntryResolver(db, gitserverClient,
		NewGitCommitResolver(db, gitserverClient, NewRepositoryResolver(db, gitserverClient, &types.Repo{Name: "my/repo"}), api.CommitID("aaaa"), nil),
		CreateFileInfo(wantPath, true))

	entry := NewTreeEntryStatsResolver(gitTree)

	langStats, err := entry.Languages(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	langStatsResult := make([]inventory.Lang, 0, len(langStats))
	for _, ls := range langStats {
		langStatsResult = append(langStatsResult, ls.l)
	}

	if diff := cmp.Diff(langStatsResult, wantLangStats); diff != "" {
		t.Errorf("langStats are wrong: %s", diff)
	}
	t.Logf("%v", langStats)
}
