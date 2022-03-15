package inference

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TypeScriptPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		pathPattern(rawPattern("tsconfig.json")),
		pathPattern(rawPattern("package.json")),
		pathPattern(rawPattern("lerna.json")),
		pathPattern(rawPattern("yarn.lock")),
		pathPattern(rawPattern(".nvmrc")),
		pathPattern(rawPattern(".node-version")),
		pathPattern(rawPattern(".n-node-version")),
	}
}

const lsifTypescriptImage = "sourcegraph/lsif-typescript:autoindex"

var tscSegmentBlockList = append([]string{"node_modules"}, segmentBlockList...)

func inferSingleTypeScriptIndexJob(
	gitclient GitClient, pathMap pathMap, tsConfigPath string, shouldInferConfig bool,
) *config.IndexJob {
	if !containsNoSegments(tsConfigPath, tscSegmentBlockList...) {
		return nil
	}
	isYarn := checkLernaFile(gitclient, tsConfigPath, pathMap)
	var dockerSteps []config.DockerStep
	for _, dir := range ancestorDirs(tsConfigPath) {
		if !pathMap.contains(dir, "package.json") {
			continue
		}

		ignoreScripts := ""
		if shouldInferConfig {
			ignoreScripts = " --ignore-scripts"
		}
		var commands []string
		if isYarn || pathMap.contains(dir, "yarn.lock") {
			commands = append(commands, "yarn --ignore-engines"+ignoreScripts)
		} else {
			commands = append(commands, "npm install"+ignoreScripts)
		}

		dockerSteps = append(dockerSteps, config.DockerStep{
			Root:     dir,
			Image:    lsifTypescriptImage,
			Commands: commands,
		})
	}

	n := len(dockerSteps)
	for i := 0; i < n/2; i++ {
		dockerSteps[i], dockerSteps[n-i-1] = dockerSteps[n-i-1], dockerSteps[i]
	}

	indexerArgs := []string{"lsif-typescript-autoindex", "index"}
	if shouldInferConfig {
		indexerArgs = append(indexerArgs, "--inferTSConfig")
	}

	return &config.IndexJob{
		Steps:       dockerSteps,
		LocalSteps:  nil,
		Root:        dirWithoutDot(tsConfigPath),
		Indexer:     lsifTypescriptImage,
		IndexerArgs: indexerArgs,
		Outfile:     "",
	}
}

func InferTypeScriptIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	pathMap := newPathMap(paths)

	tsConfigEntry, tsConfigPresent := pathMap["tsconfig.json"]
	if !tsConfigPresent {
		indexJob := inferSingleTypeScriptIndexJob(gitclient, pathMap, "tsconfig.json", true)
		if indexJob != nil {
			return []config.IndexJob{*indexJob}
		}
		return []config.IndexJob{}
	}

	for _, tsConfigIndex := range tsConfigEntry.indexes {
		indexJob := inferSingleTypeScriptIndexJob(gitclient, pathMap, paths[tsConfigIndex], false)
		if indexJob != nil {
			indexes = append(indexes, *indexJob)
		}
	}

	return indexes
}

func checkLernaFile(gitclient GitClient, path string, pathMap pathMap) (isYarn bool) {
	lernaConfig := struct {
		NpmClient string `json:"npmClient"`
	}{}

	for _, dir := range ancestorDirs(path) {
		if pathMap.contains(dir, "lerna.json") {
			lernaPath := filepath.Join(dir, "lerna.json")
			if b, err := gitclient.RawContents(context.TODO(), lernaPath); err == nil {
				if err := json.Unmarshal(b, &lernaConfig); err == nil && lernaConfig.NpmClient == "yarn" {
					return true
				}
			}
		}
	}
	return false
}
