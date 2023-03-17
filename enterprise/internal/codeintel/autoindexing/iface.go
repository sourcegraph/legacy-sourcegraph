package autoindexing

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/jobselector"
)

type DependenciesService = background.DependenciesService

type PoliciesService = background.PoliciesService

type ReposStore = background.ReposStore

type GitserverRepoStore = background.GitserverRepoStore

type ExternalServiceStore = background.ExternalServiceStore

type PolicyMatcher = background.PolicyMatcher

type RepoUpdaterClient interface {
	background.RepoUpdaterClient
	enqueuer.RepoUpdaterClient
}

type InferenceService = jobselector.InferenceService

type UploadService = background.UploadService
