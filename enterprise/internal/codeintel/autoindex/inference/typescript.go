package inference

import (
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"
)

const (
	lsifTscImage     = "sourcegraph/lsif-node:latest"
	nodeInstallImage = "node:alpine3.12"
)

type lsifTscJobRecognizer struct{}

var _ IndexJobRecognizer = lsifTscJobRecognizer{}

type lernaConfig struct {
	NPMClient string `json:"npmClient"`
}

func (r lsifTscJobRecognizer) CanIndex(paths []string, gitserver GitserverClientWrapper) bool {
	for _, path := range paths {
		if r.canIndexPath(path) {
			return true
		}
	}

	return false
}

func (r lsifTscJobRecognizer) InferIndexJobs(paths []string, gitserver GitserverClientWrapper) (indexes []IndexJob) {
	for _, path := range paths {
		if !r.canIndexPath(path) {
			continue
		}

		var isYarn bool
		var dockerSteps []DockerStep
		for _, dir := range ancestorDirs(path) {
			if exists := contains(paths, filepath.Join(dir, "lerna.json")); exists && !isYarn {
				if b, err := gitserver.RawContents(context.TODO(), "lerna.json"); err == nil {
					var c lernaConfig
					if err := json.Unmarshal(b, &c); err == nil {
						isYarn = c.NPMClient == "yarn"
					}
				}
			}

			if !contains(paths, filepath.Join(dir, "package.json")) {
				continue
			}

			var commands []string
			if isYarn || contains(paths, filepath.Join(dir, "yarn.lock")) {
				commands = append(commands, "yarn --ignore-engines")
			} else {
				commands = append(commands, "npm install")
			}

			dockerSteps = append(dockerSteps, DockerStep{
				Root:     dir,
				Image:    nodeInstallImage,
				Commands: commands,
			})
		}

		n := len(dockerSteps)
		for i := 0; i < n/2; i++ {
			dockerSteps[i], dockerSteps[n-i-1] = dockerSteps[n-i-1], dockerSteps[i]
		}

		indexes = append(indexes, IndexJob{
			DockerSteps: dockerSteps,
			Root:        dirWithoutDot(path),
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		})
	}

	return indexes
}

func (lsifTscJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("tsconfig.json"),
		suffixPattern("package.json"),
		suffixPattern("lerna.json"),
	}
}

func (r lsifTscJobRecognizer) canIndexPath(path string) bool {
	// TODO(efritz) - check for javascript files
	return filepath.Base(path) == "tsconfig.json" && containsNoSegments(path, tscSegmentBlockList...)
}

var tscSegmentBlockList = append([]string{
	"node_modules",
}, segmentBlockList...)
