package cloud

import (
	"cmp"
	"context"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/ci/gitops"
	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var ErrDeploymentExists error = errors.New("deployment already exists")

var DeployEphemeralCommand = cli.Command{
	Name:        "deploy",
	Usage:       "sg could deploy --branch <branch> --tag <tag>",
	Description: "Deploy the specified branch or tag to an ephemeral Sourcegraph Cloud environment",
	Action:      wipAction(deployCloudEphemeral),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			DefaultText: "the name of the ephemeral deployment. If not specified, the name will be derived from the branch name",
		},
		&cli.StringFlag{
			Name:        "version",
			DefaultText: "deploys an ephemeral cloud Sourcegraph environment with the specified version. The version MUST exist and implies that no build will be created",
		},
	},
}

func determineVersion(build *buildkite.Build, tag string) (string, error) {
	if tag == "" {
		t, err := gitops.GetLatestTag()
		if err != nil {
			if err != gitops.ErrNoTags {
				return "", err
				// if we get no tags then we use an empty string - this is how it is done in CI
			}
			t = ""
		}
		tag = t
	}

	return images.BranchImageTag(
		time.Now(),
		pointers.DerefZero(build.Commit),
		pointers.DerefZero(build.Number),
		pointers.DerefZero(build.Branch),
		tag,
	), nil
}

func createDeploymentForVersion(ctx context.Context, name, version string) error {
	email, err := GetGCloudAccount(ctx)
	if err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	cloudClient, err := NewClient(ctx, email, APIEndpoint)
	if err != nil {
		return err
	}

	cloudEmoji := "☁️"
	pending := std.Out.Pending(output.Linef(cloudEmoji, output.StylePending, "Starting deployment %q for version %q", name, version))
	spec := NewDeploymentSpec(
		sanitizeInstanceName(name),
		version,
	)

	inst, err := cloudClient.GetInstance(ctx, spec.Name)
	if err != nil {
		if !errors.Is(err, ErrInstanceNotFound) {
			return errors.Wrapf(err, "failed to determine if instance %q already exists", spec.Name)
		}
	} else {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Cannot deploy %q", err))
		// Deployment exists
		return ErrDeploymentExists
	}

	inst, err = cloudClient.CreateInstance(ctx, spec)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Deployment failed: %v", err))
		return errors.Wrapf(err, "failed to deploy version %v", version)
	}

	pending.Writef("Deploy instance details: \n%s", inst.String())
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Deployment %q created for version %q - access at: %s", name, version, inst.URL))
	return nil
}

func triggerEphemeralBuild(ctx context.Context, currRepo *repo.GitRepo) (*buildkite.Build, error) {
	pending := std.Out.Pending(output.Linef("🔨", output.StylePending, "Checking if branch %q is up to date with remote", currRepo.Branch))
	if isOutOfSync, err := currRepo.IsOutOfSync(ctx); err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "branch is out of date with remote"))
		return nil, err
	} else if isOutOfSync {
		return nil, ErrBranchOutOfSync
	}

	client, err := bk.NewClient(ctx, std.Out)
	if err != nil {
		return nil, err
	}

	pending.Updatef("Starting cloud ephemeral build for %q on commit %q", currRepo.Branch, currRepo.Ref)
	build, err := client.TriggerBuild(ctx, "sourcegraph", currRepo.Branch, currRepo.Ref, bk.WithEnvVar("CLOUD_EPHEMERAL", "true"))
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to trigger build"))
		return nil, err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Build %d created. Build progress can be viewed at %s", pointers.DerefZero(build.Number), pointers.DerefZero(build.WebURL)))

	return build, nil
}

func deployCloudEphemeral(ctx *cli.Context) error {
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}

	// TODO(burmudar): We need to handle tags
	var currRepo *repo.GitRepo
	// We are on the branch we want to deploy, so we use the current commit
	head, err := repo.GetHeadCommit(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current commit")
	}
	currRepo = repo.NewGitRepo(currentBranch, head)

	version := ctx.String("version")
	// if a version is specified we do not build anything and just trigger the cloud deployment
	if version == "" {
		build, err := triggerEphemeralBuild(ctx.Context, currRepo)
		if err != nil {
			if err == ErrBranchOutOfSync {
				std.Out.WriteWarningf(`Your branch %q is out of sync with remote.

Please make sure you have either pushed or pulled the latest changes before trying again`, currRepo.Branch)
			} else {
				std.Out.WriteFailuref("Cannot start deployment as there was problem with the ephemeral build")
			}
			return errors.Wrapf(err, "cloud ephemeral deployment failure")
		}

		version, err = determineVersion(build, ctx.String("tag"))
		if err != nil {
			return err
		}
	}

	name := cmp.Or(ctx.String("name"), currRepo.Branch)
	err = createDeploymentForVersion(ctx.Context, name, version)
	if err != nil {
		if errors.Is(err, ErrDeploymentExists) {
			std.Out.WriteWarningf("Cannot create a new deployment as a deployment with name %q already exists", name)
			std.Out.WriteSuggestionf(`You might want to try one of the following:
- Specify a different deployment name with the --name flag
- Upgrade the current deployment instead by using the upgrade command instead of deploy`)
		}
		return err
	}
	return nil
}
