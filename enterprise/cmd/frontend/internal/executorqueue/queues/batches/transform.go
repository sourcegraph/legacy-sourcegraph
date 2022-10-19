package batches

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	srcInputPath         = "input.json"
	srcRepoDir           = "repository"
	srcTempDir           = ".src-tmp"
	srcWorkspaceFilesDir = "workspace-files"
)

type BatchesStore interface {
	GetBatchSpecWorkspace(context.Context, store.GetBatchSpecWorkspaceOpts) (*btypes.BatchSpecWorkspace, error)
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)
	ListBatchSpecWorkspaceFiles(ctx context.Context, opts store.ListBatchSpecWorkspaceFileOpts) ([]*btypes.BatchSpecWorkspaceFile, int64, error)

	DatabaseDB() database.DB
}

const fileStoreBucket = "batch-changes"

// transformRecord transforms a *btypes.BatchSpecWorkspaceExecutionJob into an apiclient.Job.
func transformRecord(ctx context.Context, logger log.Logger, s BatchesStore, job *btypes.BatchSpecWorkspaceExecutionJob) (apiclient.Job, error) {
	workspace, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: job.BatchSpecWorkspaceID})
	if err != nil {
		return apiclient.Job{}, errors.Wrapf(err, "fetching workspace %d", job.BatchSpecWorkspaceID)
	}

	batchSpec, err := s.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: workspace.BatchSpecID})
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching batch spec")
	}

	// This should never happen. To get some easier debugging when a user sees strange
	// behavior, we log some additional context.
	if job.UserID != batchSpec.UserID {
		logger.Error("bad DB state: batch spec workspace execution job did not have the same user ID as the associated batch spec")
	}

	// Next, we fetch all secrets that are requested for the execution.
	// Use a user context here, so that only secrets are returned that the user may see at this point in time.
	userCtx := actor.WithActor(ctx, actor.FromUser(batchSpec.UserID))
	esStore := s.DatabaseDB().ExecutorSecrets(keyring.Default().ExecutorSecretKey)
	secrets, _, err := esStore.List(userCtx, database.ExecutorSecretScopeBatches, database.ExecutorSecretsListOpts{
		NamespaceUserID: batchSpec.NamespaceUserID,
		NamespaceOrgID:  batchSpec.NamespaceOrgID,
		Keys:            requiredSecrets(batchSpec.Spec),
	})
	if err != nil {
		return apiclient.Job{}, err
	}

	// And build the env vars from the secrets.
	secretEnvVars := make([]string, len(secrets))
	redactedEnvVars := make(map[string]string, len(secrets))
	for i, secret := range secrets {
		// TODO: Create audit log entry.

		val, err := secret.EncryptedValue.Decrypt(ctx)
		if err != nil {
			return apiclient.Job{}, err
		}
		ev, err := secret.EnvVar(ctx)
		if err != nil {
			return apiclient.Job{}, err
		}
		secretEnvVars[i] = ev
		redactedEnvVars[val] = fmt.Sprintf("${{ secrets.%s }}", secret.Key)
	}

	// 🚨 SECURITY: Set the actor on the context so we check for permissions
	// when loading the repository.
	ctx = actor.WithActor(ctx, actor.FromUser(job.UserID))

	repo, err := s.DatabaseDB().Repos().Get(ctx, workspace.RepoID)
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching repo")
	}

	// TODO: For now, until we can figure out which secrets are actually required:
	// Load all global secrets and all secrets for the namespace this batch spec
	// lives in, then overlay these, with org having higher precedence.

	executionInput := batcheslib.WorkspacesExecutionInput{
		Repository: batcheslib.WorkspaceRepo{
			ID:   string(graphqlbackend.MarshalRepositoryID(repo.ID)),
			Name: string(repo.Name),
		},
		Branch: batcheslib.WorkspaceBranch{
			Name:   workspace.Branch,
			Target: batcheslib.Commit{OID: workspace.Commit},
		},
		Path:               workspace.Path,
		OnlyFetchWorkspace: workspace.OnlyFetchWorkspace,
		// TODO: We can further optimize here later and tell src-cli to
		// not run those steps so there is no discrepancy between the backend
		// and src-cli calculating the if conditions.
		Steps:             batchSpec.Spec.Steps,
		SearchResultPaths: workspace.FileMatches,
		BatchChangeAttributes: template.BatchChangeAttributes{
			Name:        batchSpec.Spec.Name,
			Description: batchSpec.Spec.Description,
		},
	}

	// Check if we have a cache result for the workspace, if so, add it to the execution
	// input.
	if !batchSpec.NoCache {
		// Find the cache entry for the _last_ step. src-cli only needs the most
		// recent cache entry to do its work.
		latestStepIndex := -1
		for stepIndex := range workspace.StepCacheResults {
			if stepIndex > latestStepIndex {
				latestStepIndex = stepIndex
			}
		}
		if latestStepIndex != -1 {
			cacheEntry, ok := workspace.StepCacheResult(latestStepIndex)
			// Technically this should never be not ok, but computers.
			if ok {
				executionInput.CachedStepResultFound = true
				executionInput.CachedStepResult = *cacheEntry.Value
			}
		}
	}

	// Marshal the execution input into JSON and add it to the files passed to
	// the VM.
	marshaledInput, err := json.Marshal(executionInput)
	if err != nil {
		return apiclient.Job{}, err
	}
	files := map[string]apiclient.VirtualMachineFile{
		srcInputPath: {
			Content: string(marshaledInput),
		},
	}

	workspaceFiles, _, err := s.ListBatchSpecWorkspaceFiles(ctx, store.ListBatchSpecWorkspaceFileOpts{BatchSpecRandID: batchSpec.RandID})
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching workspace files")
	}
	for _, workspaceFile := range workspaceFiles {
		files[filepath.Join(srcWorkspaceFilesDir, workspaceFile.Path, workspaceFile.FileName)] = apiclient.VirtualMachineFile{
			Bucket:     fileStoreBucket,
			Key:        filepath.Join(batchSpec.RandID, workspaceFile.RandID),
			ModifiedAt: workspaceFile.ModifiedAt,
		}
	}

	// If we only want to fetch the workspace, we add a sparse checkout pattern.
	sparseCheckout := []string{}
	if workspace.OnlyFetchWorkspace {
		sparseCheckout = []string{
			fmt.Sprintf("%s/*", workspace.Path),
		}
	}

	commands := []string{
		"batch",
		"exec",
		"-f", srcInputPath,
		"-repo", srcRepoDir,
		// Tell src to store tmp files inside the workspace. Src currently
		// runs on the host and we don't want pollution outside of the workspace.
		"-tmp", srcTempDir,
	}
	// Only add the workspaceFiles flag if there are files to mount. This helps with backwards compatibility.
	if len(workspaceFiles) > 0 {
		commands = append(commands, "-workspaceFiles", srcWorkspaceFilesDir)
	}

	j := apiclient.Job{
		ID:                  int(job.ID),
		VirtualMachineFiles: files,
		RepositoryName:      string(repo.Name),
		RepositoryDirectory: srcRepoDir,
		Commit:              workspace.Commit,
		// We only care about the current repos content, so a shallow clone is good enough.
		// Later we might allow to tweak more git parameters, like submodules and LFS.
		ShallowClone:   true,
		SparseCheckout: sparseCheckout,
		CliSteps: []apiclient.CliStep{
			{
				Key:      "batch-exec",
				Commands: commands,
				Dir:      ".",
				Env:      secretEnvVars,
			},
		},
		RedactedValues: redactedEnvVars,
	}

	fmt.Printf("Job payload!!!!:\n%+v\n\n", j)

	return j, nil
}

// requiredSecrets inspects all steps for env vars used and compiles a deduplicated
// list from those.
func requiredSecrets(spec *batcheslib.BatchSpec) []string {
	requiredSecretsMap := map[string]struct{}{}
	requiredSecrets := []string{}
	for _, step := range spec.Steps {
		for _, v := range step.Env.OuterVars() {
			if _, ok := requiredSecretsMap[v]; !ok {
				requiredSecretsMap[v] = struct{}{}
				requiredSecrets = append(requiredSecrets, v)
			}
		}
	}
	return requiredSecrets
}
