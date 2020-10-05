package inference

import (
	"path/filepath"
	"regexp"
)

type lsifTscJobRecognizer struct{}

var _ IndexJobRecognizer = lsifTscJobRecognizer{}

func (lsifTscJobRecognizer) CanIndex(paths []string) bool {
	for _, path := range paths {
		if filepath.Base(path) == "tsconfig.json" && !containsSegment(path, "node_modules") {
			return true
		}

		// TODO(efritz) - check for javascript files
	}

	return false
}

func (lsifTscJobRecognizer) InferIndexJobs(paths []string) (indexes []IndexJob) {
	for _, path := range paths {
		if filepath.Base(path) == "tsconfig.json" && !containsSegment(path, "node_modules") {
			var dockerSteps []DockerStep
			for _, dir := range ancestorDirs(path) {
				if contains(paths, filepath.Join(dir, "yarn.lock")) {
					dockerSteps = append(dockerSteps, DockerStep{
						Root:     dir,
						Image:    "node:alpine3.12",
						Commands: []string{"yarn"},
					})

					break
				}

				if contains(paths, filepath.Join(dir, "package.json")) {
					dockerSteps = append(dockerSteps, DockerStep{
						Root:     dir,
						Image:    "node:alpine3.12",
						Commands: []string{"npm", "install"},
					})

					break
				}
			}

			indexes = append(indexes, IndexJob{
				DockerSteps: dockerSteps,
				Root:        dirWithoutDot(path),
				Indexer:     "sourcegraph/lsif-node:latest",
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			})
		}
	}

	return indexes
}

func (lsifTscJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("tsconfig.json"),
		suffixPattern("package.json"),
	}
}
