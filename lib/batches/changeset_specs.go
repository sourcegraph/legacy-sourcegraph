package batches

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

var errOptionalPublishedUnsupported = NewValidationError(errors.New(`This Sourcegraph version requires the "published" field to be specified in the batch spec; upgrade to version 3.30.0 or later to be able to omit the published field and control publication from the UI.`))

type ChangesetSpecRepository struct {
	Name        string
	FileMatches []string

	BaseRef string
	BaseRev string
}
type ChangesetSpecInput struct {
	// This is the old graphql.Repository
	BaseRepositoryID string
	HeadRepositoryID string

	// This is just a thin wrapper
	Repository ChangesetSpecRepository

	// These were on Task
	BatchChangeAttributes *template.BatchChangeAttributes `json:"-"`
	Template              *ChangesetTemplate              `json:"-"`
	TransformChanges      *TransformChanges               `json:"-"`

	// These were on executionResult

	// Diff is the produced by executing all steps.
	Diff string `json:"diff"`
	// ChangedFiles are files that have been changed by all steps.
	ChangedFiles *git.Changes `json:"changedFiles"`
	// Outputs are the outputs produced by all steps.
	Outputs map[string]interface{} `json:"outputs"`
	// Path relative to the repository's root directory in which the steps
	// have been executed.
	// No leading slashes. Root directory is blank string.
	Path string
}

type ChangesetSpecFeatureFlags struct {
	IncludeAutoAuthorDetails bool
	AllowOptionalPublished   bool
}

func BuildChangesetSpecs(input *ChangesetSpecInput, features ChangesetSpecFeatureFlags) ([]*ChangesetSpec, error) {
	tmplCtx := &template.ChangesetTemplateContext{
		BatchChangeAttributes: *input.BatchChangeAttributes,
		Steps: template.StepsContext{
			Changes: input.ChangedFiles,
			Path:    input.Path,
		},
		Outputs: input.Outputs,
		Repository: template.Repository{
			Name:        input.Repository.Name,
			FileMatches: input.Repository.FileMatches,
		},
	}

	var authorName string
	var authorEmail string

	if input.Template.Commit.Author == nil {
		if features.IncludeAutoAuthorDetails {
			// user did not provide author info, so use defaults
			authorName = "Sourcegraph"
			authorEmail = "batch-changes@sourcegraph.com"
		}
	} else {
		var err error
		authorName, err = template.RenderChangesetTemplateField("authorName", input.Template.Commit.Author.Name, tmplCtx)
		if err != nil {
			return nil, err
		}
		authorEmail, err = template.RenderChangesetTemplateField("authorEmail", input.Template.Commit.Author.Email, tmplCtx)
		if err != nil {
			return nil, err
		}
	}

	title, err := template.RenderChangesetTemplateField("title", input.Template.Title, tmplCtx)
	if err != nil {
		return nil, err
	}

	body, err := template.RenderChangesetTemplateField("body", input.Template.Body, tmplCtx)
	if err != nil {
		return nil, err
	}

	message, err := template.RenderChangesetTemplateField("message", input.Template.Commit.Message, tmplCtx)
	if err != nil {
		return nil, err
	}

	// TODO: As a next step, we should extend the ChangesetTemplateContext to also include
	// TransformChanges.Group and then change validateGroups and groupFileDiffs to, for each group,
	// render the branch name *before* grouping the diffs.
	defaultBranch, err := template.RenderChangesetTemplateField("branch", input.Template.Branch, tmplCtx)
	if err != nil {
		return nil, err
	}

	newSpec := func(branch, diff string) (*ChangesetSpec, error) {
		var published interface{} = nil
		if input.Template.Published != nil {
			published = input.Template.Published.ValueWithSuffix(input.Repository.Name, branch)

			// Backward compatibility: before optional published fields were
			// allowed, ValueWithSuffix() would fall back to false, not nil. We
			// need to replicate this behaviour here.
			if published == nil && !features.AllowOptionalPublished {
				published = false
			}
		} else if !features.AllowOptionalPublished {
			return nil, errOptionalPublishedUnsupported
		}

		return &ChangesetSpec{
			BaseRepository: input.BaseRepositoryID,
			HeadRepository: input.HeadRepositoryID,
			BaseRef:        input.Repository.BaseRef,
			BaseRev:        input.Repository.BaseRev,

			HeadRef: ensureRefPrefix(branch),
			Title:   title,
			Body:    body,
			Commits: []GitCommitDescription{
				{
					Message:     message,
					AuthorName:  authorName,
					AuthorEmail: authorEmail,
					Diff:        diff,
				},
			},
			Published: PublishedValue{Val: published},
		}, nil
	}

	var specs []*ChangesetSpec

	groups := groupsForRepository(input.Repository.Name, input.TransformChanges)
	if len(groups) != 0 {
		err := validateGroups(input.Repository.Name, input.Template.Branch, groups)
		if err != nil {
			return specs, err
		}

		// TODO: Regarding 'defaultBranch', see comment above
		diffsByBranch, err := groupFileDiffs(input.Diff, defaultBranch, groups)
		if err != nil {
			return specs, errors.Wrap(err, "grouping diffs failed")
		}

		for branch, diff := range diffsByBranch {
			spec, err := newSpec(branch, diff)
			if err != nil {
				return nil, err
			}
			specs = append(specs, spec)
		}
	} else {
		spec, err := newSpec(defaultBranch, input.Diff)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

func groupsForRepository(repoName string, transform *TransformChanges) []Group {
	groups := []Group{}

	if transform == nil {
		return groups
	}

	for _, g := range transform.Group {
		if g.Repository != "" {
			if g.Repository == repoName {
				groups = append(groups, g)
			}
		} else {
			groups = append(groups, g)
		}
	}

	return groups
}

func validateGroups(repoName, defaultBranch string, groups []Group) error {
	uniqueBranches := make(map[string]struct{}, len(groups))

	for _, g := range groups {
		if _, ok := uniqueBranches[g.Branch]; ok {
			return NewValidationError(fmt.Errorf("transformChanges would lead to multiple changesets in repository %s to have the same branch %q", repoName, g.Branch))
		} else {
			uniqueBranches[g.Branch] = struct{}{}
		}

		if g.Branch == defaultBranch {
			return NewValidationError(fmt.Errorf("transformChanges group branch for repository %s is the same as branch %q in changesetTemplate", repoName, defaultBranch))
		}
	}

	return nil
}

func groupFileDiffs(completeDiff, defaultBranch string, groups []Group) (map[string]string, error) {
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(completeDiff))
	if err != nil {
		return nil, err
	}

	// Housekeeping: we setup these two datastructures so we can
	// - access the group.Branch by the directory for which they should be used
	// - check against the given directories, in order.
	branchesByDirectory := make(map[string]string, len(groups))
	dirs := make([]string, len(branchesByDirectory))
	for _, g := range groups {
		branchesByDirectory[g.Directory] = g.Branch
		dirs = append(dirs, g.Directory)
	}

	byBranch := make(map[string][]*diff.FileDiff, len(groups))
	byBranch[defaultBranch] = []*diff.FileDiff{}

	// For each file diff...
	for _, f := range fileDiffs {
		name := f.NewName
		if name == "/dev/null" {
			name = f.OrigName
		}

		// .. we check whether it matches one of the given directories in the
		// group transformations, with the last match winning:
		var matchingDir string
		for _, d := range dirs {
			if strings.Contains(name, d) {
				matchingDir = d
			}
		}

		// If the diff didn't match a rule, it goes into the default branch and
		// the default changeset.
		if matchingDir == "" {
			byBranch[defaultBranch] = append(byBranch[defaultBranch], f)
			continue
		}

		// If it *did* match a directory, we look up which branch we should use:
		branch, ok := branchesByDirectory[matchingDir]
		if !ok {
			panic("this should not happen: " + matchingDir)
		}

		byBranch[branch] = append(byBranch[branch], f)
	}

	finalDiffsByBranch := make(map[string]string, len(byBranch))
	for branch, diffs := range byBranch {
		printed, err := diff.PrintMultiFileDiff(diffs)
		if err != nil {
			return nil, errors.Wrap(err, "printing multi file diff failed")
		}
		finalDiffsByBranch[branch] = string(printed)
	}
	return finalDiffsByBranch, nil
}

func ensureRefPrefix(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return ref
	}
	return "refs/heads/" + ref
}
