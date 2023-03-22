package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

type UploadService interface {
	GetDirtyRepositories(ctx context.Context) (_ []shared.DirtyRepository, err error)
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
}

type Locker = background.Locker

type RepoStore = background.RepoStore

type PolicyService = background.PolicyService

type PolicyMatcher = background.PolicyMatcher
