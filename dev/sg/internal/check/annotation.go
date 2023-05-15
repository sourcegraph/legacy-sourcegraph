package check

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func generateAnnotation(category string, check string, content string) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return // do nothing
	}

	// set up annotations dir
	annotationsDir := filepath.Join(repoRoot, "annotations")
	os.MkdirAll(annotationsDir, os.ModePerm)

	// write annotation
	path := filepath.Join(annotationsDir, fmt.Sprintf("%s: %s.md", category, check))
	_ = os.WriteFile(path, []byte(content+"\n"), os.ModePerm)

	if check == "Go format" {
		gofmt, _ := os.OpenFile(fmt.Sprintf("%s/gofmt", annotationsDir))
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		defer gofmt.Close()

		// Read the contents of the source file
		contents, err := ioutil.ReadAll(gofmt)
		if err != nil {
			panic(err)
		}
		annotationFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		defer annotationFile.Close()

		_, err = annotationFile.WriteString(string(contents))
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}

		_ = os.Remove(fmt.Sprintf("%s/gofmt", annotationsDir))

	}
}
