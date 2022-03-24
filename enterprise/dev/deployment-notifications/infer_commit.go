package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"
)

// git diff ...
// +         image: index.docker.io/sourcegraph/migrator:137540_2022-03-17_d24138504aea@sha256:2b6efe8f447b22f9396544f885f2f326d21325d652f9b36961f3d105723789df
// -         image: index.docker.io/sourcegraph/migrator:137540_2022-03-17_XXXXXXXXXXXX@sha256:2b6efe8f447b22f9396544f885f2f326d21325d652f9b36961f3d105723789df
var imageCommitRegexp = `(?m)^DIFF_OP\s+image:\s[^/]+\/sourcegraph\/[^:]+:\d{6}_\d{4}-\d{2}-\d{2}_([^@]+)@sha256.*$` // (?m) stands for multiline.

type manifestDeploymentDiffer struct {
	changedFiles []string
	diffs        map[string]*ApplicationVersionDiff
}

type mockDeploymentDiffer struct {
	diffs map[string]*ApplicationVersionDiff
}

func (m *mockDeploymentDiffer) Applications() (map[string]*ApplicationVersionDiff, error) {
	return m.diffs, nil
}

func NewManifestDeploymentDiffer(basedir string, changedFiles []string) DeploymentDiffer {
	return &manifestDeploymentDiffer{
		changedFiles: changedFiles,
	}
}

func NewMockManifestDeployementsDiffer(m map[string]*ApplicationVersionDiff) DeploymentDiffer {
	return &mockDeploymentDiffer{
		diffs: m,
	}
}

func (m *manifestDeploymentDiffer) Applications() (map[string]*ApplicationVersionDiff, error) {
	err := m.parseManifests()
	if err != nil {
		return nil, err
	}
	return m.diffs, nil
}

func (m *manifestDeploymentDiffer) parseManifests() error {
	apps := map[string]*ApplicationVersionDiff{}
	for _, path := range m.changedFiles {
		fmt.Println(path)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			// If the file is a directory, skip it.
			continue
		}
		elems := strings.Split(path, string(filepath.Separator))
		if len(elems) < 1 {
			// If the file is at the root, skip it. Applications are always in subfolders.
			continue
		}
		appName := elems[1] // base/elems[1]/...

		filename := filepath.Base(path)
		components := strings.Split(filename, ".")
		if len(components) < 3 {
			// If the file isn't name like appName.Kind.yaml, skip it.
			continue
		}
		kind := components[1]
		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			appDiff, err := diffDeploymentManifest(path, appName)
			if err != nil {
				return err
			}
			if appDiff != nil {
				// It's possible that we find changes that are not bumping the image, when
				// updating environment vars for example. In that case, we don't want to
				// include them.
				apps[appName] = appDiff
			}
		}
	}
	m.diffs = apps
	return nil
}

// imageDiffRegexp returns a regexp that matches an addition or deletion of an
// image tag in the image field in the manifest of an application.
func imageDiffRegexp(addition bool) *regexp.Regexp {
	var escapedOp string
	if addition {
		// If matching an addition, the + needs to be escaped to not be parsed as a
		// count operator.
		escapedOp = "\\+"
	} else {
		escapedOp = "-"
	}

	re := strings.ReplaceAll(imageCommitRegexp, "DIFF_OP", escapedOp)
	return regexp.MustCompile(re)
}

// parseSourcegraphCommitFromDeploymentManifestsDiff parses the diff output, returning
// the new and old commits that were used to build this specific image.
func parseSourcegraphCommitFromDeploymentManifestsDiff(output []byte, appname string) (*ApplicationVersionDiff, error) {
	var diff ApplicationVersionDiff
	addRegexp := imageDiffRegexp(true)
	delRegexp := imageDiffRegexp(false)

	outStr := string(output)
	matches := addRegexp.FindStringSubmatch(outStr)
	if len(matches) > 1 {
		diff.New = matches[1]
	}
	matches = delRegexp.FindStringSubmatch(outStr)
	if len(matches) > 1 {
		diff.Old = matches[1]
	}

	if diff.Old == "" || diff.New == "" {
		return nil, nil
	}

	return &diff, nil
}

func diffDeploymentManifest(path string, appName string) (*ApplicationVersionDiff, error) {
	diffCommand := []string{"diff", "@^", path}
	output, err := exec.Command("git", diffCommand...).Output()
	if err != nil {
		return nil, err
	}
	imageDiff, err := parseSourcegraphCommitFromDeploymentManifestsDiff(output, appName)
	if err != nil {
		return nil, err
	}
	return imageDiff, nil
}
