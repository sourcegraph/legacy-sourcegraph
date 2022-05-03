package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bitfield/script"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var allLintTargets = lintTargets{
	{
		Name: "urls",
		Help: "Check for broken urls in the codebase",
		Linters: []lint.Runner{
			lint.RunScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name: "go",
		Help: "Check go code for linting errors, forbidden imports, generated files, etc",
		Linters: []lint.Runner{
			lint.RunScript("Go format", "dev/check/gofmt.sh"),
			lintGoGenerate,
			lint.RunScript("Go lint", "dev/check/go-lint.sh"),
			lint.RunScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh"),
			lint.RunScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh"),
		},
	},
	{
		Name: "logger-migration",
		Help: "Run linter that enforces the new logger library",
		Linters: []lint.Runner{
			lintLoggingLibraries(),
		},
	},
	{
		Name: "docsite",
		Help: "Check the code powering docs.sourcegraph.com for broken links and linting errors",
		Linters: []lint.Runner{
			lint.RunScript("Docsite lint", "dev/check/docsite.sh"),
		},
	},
	{
		Name: "docker",
		Help: "Check Dockerfiles for Sourcegraph best practices",
		Linters: []lint.Runner{
			lint.RunScript("Docker lint", "dev/check/docker-lint.sh"),
			lintDockerfiles(),
		},
	},
	{
		Name: "client",
		Help: "Check client code for linting errors, forbidden imports, etc",
		Linters: []lint.Runner{
			lint.RunScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh"),
			lint.RunScript("Inline templates", "dev/check/template-inlines.sh"),
			lint.RunScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			lint.RunScript("SVG Compression", "dev/check/svgo.sh"),
		},
	},
	{
		Name: "shell",
		Help: "Check shell code for linting errors, formatting, etc",
		Linters: []lint.Runner{
			lint.RunScript("Shell formatting", "dev/check/shfmt.sh"),
			lint.RunScript("Shell lint", "dev/check/shellcheck.sh"),
		},
	},
}

func lintLoggingLibraries() lint.Runner {
	const header = "Logging library linter"

	return func(ctx context.Context) *lint.Report {
		start := time.Now()

		// Diffing against origin/main means that this check doesn't do much in CI, but
		// Robert can't come up with a better paradigm for this so... maybe eventually
		// we'll introduce a parameter. But for now introducing a paremeter seems more
		// confusing because it would work inconsistently across the various linters.
		diffOutput, err := script.Exec("git diff origin/main -- **/*.go").
			MatchRegexp(regexp.MustCompile(`^\+.*`)).
			String()
		if err != nil {
			return &lint.Report{
				Duration: time.Since(start),
				Header:   header,
				Output:   err.Error(),
				Err:      err,
			}
		}

		diffs := map[string][]string{}
		var currentFile string
		for _, line := range strings.Split(diffOutput, "\n") {
			if strings.HasPrefix(line, "+++") {
				currentFile = strings.Split(line, " ")[1]
			} else if len(line) > 0 {
				diffs[currentFile] = append(diffs[currentFile], strings.TrimPrefix(line, "+"))
			}
		}

		banList := []string{
			`"log"`,

			`"github.com/inconshreveable/log15"`,
			`log15.`,

			`"go.uber.org/zap"`,
			`"go.uber.org/zap/zapcore"`,
			`zap.`,
		}

		checkFile := func(f string, ls []string) error {
			for _, l := range ls {
				for _, banned := range banList {
					if strings.Contains(l, banned) {
						return errors.Newf("%s: banned usage '%s': use github.com/sourcegraph/sourcegraph/lib/log instead",
							f, banned)
					}
				}
			}
			return nil
		}

		var errs error
		for file, lines := range diffs {
			if err := checkFile(file, lines); err != nil {
				errs = errors.Append(errs, err)
			}
		}

		return &lint.Report{
			Duration: time.Since(start),
			Header:   header,
			Output: func() string {
				if errs != nil {
					return strings.TrimSpace(errs.Error())
				}
				return ""
			}(),
			Err: errs,
		}
	}
}

// lintDockerfiles runs custom Sourcegraph Dockerfile linters
func lintDockerfiles() lint.Runner {
	return func(ctx context.Context) *lint.Report {
		start := time.Now()
		var combinedErrors error
		for _, dir := range []string{
			"docker-images",
			// cmd dirs
			"cmd",
			"enterprise/cmd",
			"internal/cmd",
			// dev dirs
			"dev",
			"enterprise/dev",
		} {
			if err := filepath.Walk(dir,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !strings.Contains(filepath.Base(path), "Dockerfile") {
						return nil
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					if err := docker.ProcessDockerfile(data, docker.LintDockerfile(path)); err != nil {
						// track error but don't exit
						combinedErrors = errors.Append(combinedErrors, err)
					}

					return nil
				},
			); err != nil {
				combinedErrors = errors.Append(combinedErrors, err)
			}
		}
		return &lint.Report{
			Duration: time.Since(start),
			Header:   "Sourcegraph Dockerfile linters",
			Output: func() string {
				if combinedErrors != nil {
					return strings.TrimSpace(combinedErrors.Error())
				}
				return ""
			}(),
			Err: combinedErrors,
		}
	}
}

func lintGoGenerate(ctx context.Context) *lint.Report {
	start := time.Now()
	report := golang.Generate(ctx, nil, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header:   "Go generate check",
			Duration: report.Duration + time.Since(start),
			Err:      report.Err,
		}
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "--exit-code", "--", ".", ":!go.sum")
	out, err := cmd.CombinedOutput()
	r := lint.Report{
		Header:   "Go generate check",
		Duration: report.Duration + time.Since(start),
	}
	if err != nil {
		var sb strings.Builder
		reportOut := output.NewOutput(&sb, output.OutputOpts{
			ForceColor: true,
			ForceTTY:   true,
		})
		reportOut.WriteLine(output.Line(output.EmojiFailure, output.StyleWarning, "Uncommitted changes found after running go generate:"))
		sb.WriteString("\n")
		sb.WriteString(string(out))
		r.Err = err
		r.Output = sb.String()
		return &r
	}

	return &r
}
