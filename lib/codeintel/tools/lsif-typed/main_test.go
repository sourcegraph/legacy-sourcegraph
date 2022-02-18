package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	lsiftypedtesting "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif-typed-testing"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	reproLang "github.com/sourcegraph/sourcegraph/lib/codeintel/repro_lang/bindings/golang"
)

func TestLsifTyped(t *testing.T) {
	lsiftypedtesting.SnapshotTest(t, func(inputDirectory, outputDirectory string, sources []*lsif_typed.SourceFile) []*lsif_typed.SourceFile {
		testName := filepath.Base(inputDirectory)
		var dependencies []*reproLang.Dependency
		rootDirectory := filepath.Dir(inputDirectory)
		dirs, err := os.ReadDir(rootDirectory)
		if err != nil {
			t.Fatal(err)
		}
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			if dir.Name() == testName {
				continue
			}
			dependencyRoot := filepath.Join(rootDirectory, dir.Name())
			dependencySources, err := lsif_typed.NewSourcesFromDirectory(dependencyRoot)
			if err != nil {
				t.Fatal(err)
			}
			dependencies = append(dependencies, &reproLang.Dependency{
				Package: &lsif_typed.Package{
					Manager: "repro_manager",
					Name:    dir.Name(),
					Version: "1.0.0",
				},
				Sources: dependencySources,
			})
		}
		index, err := reproLang.Index("file:/"+inputDirectory, testName, sources, dependencies)
		if err != nil {
			t.Fatal(err)
		}
		symbolFormatter := lsif_typed.DescriptorOnlyFormatter
		symbolFormatter.IncludePackageName = func(name string) bool { return name != testName }
		snapshots, err := lsiftypedtesting.FormatSnapshots(index, "#", symbolFormatter)
		if err != nil {
			t.Fatal(err)
		}
		index.Metadata.ProjectRoot = "file:/root"
		lsif, err := reader.ConvertTypedIndexToGraphIndex(index)
		if err != nil {
			t.Fatal(err)
		}
		var obtained bytes.Buffer
		err = reader.WriteNDJSON(reader.ElementsToEmptyInterfaces(lsif), &obtained)
		if err != nil {
			t.Fatal(err)
		}
		snapshots = append(snapshots, lsif_typed.NewSourceFile(
			filepath.Join(outputDirectory, "dump.lsif"),
			"dump.lsif",
			obtained.String(),
		))
		return snapshots
	})
}
