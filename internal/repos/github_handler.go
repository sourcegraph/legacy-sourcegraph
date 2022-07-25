package repos

import (
	"context"
	"reflect"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubWebhookAPI struct {
	client *github.V3Client
}

func NewGitHubWebhookAPI(client *github.V3Client) *GitHubWebhookAPI {
	return &GitHubWebhookAPI{client: client}
}

func (g *GitHubWebhookAPI) Register(router *webhooks.GitHubWebhook) {
	router.Register(g.handleGitHubWebhook, "push")
}

func (g *GitHubWebhookAPI) handleGitHubWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Errorf("expected GitHub.PushEvent, got %s", reflect.TypeOf(event))
	}

	repoName := *event.Repo.URL
	name := api.RepoName(repoName[8:])
	repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, name)

	return nil
}

func (g *GitHubWebhookAPI) CreateSyncWebhook(ctx context.Context, repoName, targetURL, secret string) (int, error) {
	return g.client.CreateSyncWebhook(ctx, repoName, targetURL, secret)
}

func (g *GitHubWebhookAPI) ListSyncWebhooks(ctx context.Context, repoName string) ([]github.Payload, error) {
	return g.client.ListSyncWebhooks(ctx, repoName)
}

func (g *GitHubWebhookAPI) FindSyncWebhook(ctx context.Context, repoName string) (int, bool) {
	return g.client.FindSyncWebhook(ctx, repoName)
}

func (g *GitHubWebhookAPI) DeleteSyncWebhook(ctx context.Context, repoName string, hookID int) (bool, error) {
	return g.client.DeleteSyncWebhook(ctx, repoName, hookID)
}

func (g *GitHubWebhookAPI) TestPushSyncWebhook(ctx context.Context, repoName string, hookID int) (bool, error) {
	return g.client.TestPushSyncWebhook(ctx, repoName, hookID)
}
