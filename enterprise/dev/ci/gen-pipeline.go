// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/changed"
)

var preview bool
var wantYaml bool
var docs bool

func init() {
	flag.BoolVar(&preview, "preview", false, "Preview the pipeline steps")
	flag.BoolVar(&wantYaml, "yaml", false, "Use YAML instead of JSON")
	flag.BoolVar(&docs, "docs", false, "Render generated documentation")
}

func main() {
	flag.Parse()

	if docs {
		renderPipelineDocs(os.Stdout)
		return
	}

	config := ci.NewConfig(time.Now())

	pipeline, err := ci.GeneratePipeline(config)
	if err != nil {
		panic(err)
	}

	if preview {
		previewPipeline(os.Stdout, config, pipeline)
		return
	}

	if wantYaml {
		_, err = pipeline.WriteYAMLTo(os.Stdout)
	} else {
		_, err = pipeline.WriteJSONTo(os.Stdout)
	}
	if err != nil {
		panic(err)
	}
}

func previewPipeline(w io.Writer, c ci.Config, pipeline *buildkite.Pipeline) {
	fmt.Fprintf(w, "Detected run type:\n\t%s\n", c.RunType.String())
	fmt.Fprintf(w, "Detected diffs:\n\t%s\n", c.Diff.String())
	fmt.Fprintf(w, "Computed build steps:\n")
	printPipeline(w, "", pipeline)
}

func printPipeline(w io.Writer, prefix string, pipeline *buildkite.Pipeline) {
	if pipeline.Group.Group != "" {
		fmt.Fprintf(w, "%s%s\n", prefix, pipeline.Group.Group)
	}
	for _, raw := range pipeline.Steps {
		switch v := raw.(type) {
		case *buildkite.Step:
			printStep(w, prefix, v)
		case *buildkite.Pipeline:
			printPipeline(w, prefix+"\t", v)
		}
	}
}

func printStep(w io.Writer, prefix string, step *buildkite.Step) {
	fmt.Fprintf(w, "%s\t%s\n", prefix, step.Label)
	switch {
	case len(step.DependsOn) > 5:
		fmt.Fprintf(w, "%s\t\t→ depends on %s, ... (%d more steps)\n", prefix, strings.Join(step.DependsOn[0:5], ", "), len(step.DependsOn)-5)
	case len(step.DependsOn) > 0:
		fmt.Fprintf(w, "%s\t\t→ depends on %s\n", prefix, strings.Join(step.DependsOn, " "))
	}
}

var emojiRegexp = regexp.MustCompile(`:(\S*):`)

func trimEmoji(s string) string {
	return strings.TrimSpace(emojiRegexp.ReplaceAllString(s, ""))
}

func renderPipelineDocs(w io.Writer) {
	fmt.Fprintln(w, "## Run types")

	// Introduce pull request builds first
	fmt.Fprintf(w, "\n### %s\n", ci.PullRequest.String())
	fmt.Fprintln(w, "\nThe default run type.")

	for diff := changed.Go; diff < changed.All; diff <<= 1 {
		pipeline, err := ci.GeneratePipeline(ci.Config{
			RunType: ci.PullRequest,
			Diff:    diff,
		})
		if err != nil {
			log.Fatalf("Generating pipeline for diff %q: %s", diff, err)
		}
		fmt.Fprintf(w, "\n- **With %q changes:**\n", diff)
		for _, raw := range pipeline.Steps {
			switch v := raw.(type) {
			case *buildkite.Step:
				fmt.Fprintf(w, "  - %s\n", trimEmoji(v.Label))
			case *buildkite.Pipeline:
				var steps []string
				for _, step := range v.Steps {
					s, ok := step.(*buildkite.Step)
					if ok {
						steps = append(steps, trimEmoji(s.Label))
					}
				}
				fmt.Fprintf(w, "  - %s: %s\n", v.Group.Group, strings.Join(steps, ", "))

			}
		}
	}

	// Introduce the others
	for rt := ci.PullRequest + 1; rt < ci.None; rt += 1 {
		fmt.Fprintf(w, "\n### %s\n", rt.String())
		if m := rt.Matcher(); m == nil {
			fmt.Fprintln(w, "\nNo matcher defined")
		} else {
			if m.Branch != "" {
				matchName := fmt.Sprintf("`%s`", m.Branch)
				if m.BranchRegexp {
					matchName += " (regexp)"
				}
				if m.BranchExact {
					matchName += " (exact)"
				}
				fmt.Fprintf(w, "\n- Branches matching %s", matchName)
			}
			if m.TagPrefix != "" {
				fmt.Fprintf(w, "\n- Tags starting with `%s`", m.TagPrefix)
			}
			if len(m.EnvIncludes) > 0 {
				fmt.Fprintf(w, "\n- Environment including `%+v`", m.EnvIncludes)
			}
			fmt.Fprintln(w)
		}
	}
}
