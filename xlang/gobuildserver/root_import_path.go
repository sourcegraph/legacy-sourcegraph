package gobuildserver

import (
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/ctxvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

// determineRootImportPath determines the root import path for the Go
// workspace. It looks at canonical import path comments and the
// repo's original clone URL to infer it.
//
// It's intended to handle cases like
// github.com/kubernetes/kubernetes, which has doc.go files that
// indicate its root import path is k8s.io/kubernetes.
func determineRootImportPath(ctx context.Context, originalRootPath string, fs ctxvfs.FileSystem) (rootImportPath string, err error) {
	if originalRootPath == "" {
		return "", errors.New("unable to determine Go workspace root import path without due to empty root path")
	}
	u, err := uri.Parse(originalRootPath)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "git":
		// TODO(keegancsmith) umami has .git paths in their import
		// paths. This normalization may in fact be incorrect. This
		// codeblock is to test if we really need this stripping in
		// production.
		if strings.HasSuffix(u.Path, ".git") {
			pathHasGitSuffix.Inc()
			defer func() {
				log.Printf("WARN: determineRootImportPath has .git suffix. before=%q after=%q %s", originalRootPath, rootImportPath, err)
			}()
		}
		rootImportPath = path.Join(u.Host, strings.TrimSuffix(u.Path, ".git"), u.FilePath())
	default:
		return "", fmt.Errorf("unrecognized originalRootPath: %q", u)
	}

	// Glide provides a canonical import path for us, try that first if it
	// exists.
	yml, err := ctxvfs.ReadFile(ctx, fs, "/glide.yaml")
	if err == nil && len(yml) > 0 {
		glide := struct {
			Package string `yaml:"package"`
			// There are other fields, but we don't use them
		}{}
		// best effort, so ignore error if we have a badly formatted
		// yml file
		_ = yaml.Unmarshal(yml, &glide)
		if glide.Package != "" {
			return glide.Package, nil
		}
	}

	// Now scan for canonical import path comments. This is a
	// heuristic; it is not guaranteed to produce the right result
	// (e.g., you could have multiple files with different canonical
	// import path comments that don't share a prefix, which is weird
	// and would break this).
	//
	// Since we have not yet set h.FS, we need to use the passed in fs.
	w := ctxvfs.Walk(ctx, "/", fs)
	const maxSlashes = 3 // heuristic, shouldn't need to traverse too deep to find this out
	const maxFiles = 25  // heuristic, shouldn't need to read too many files to find this out
	numFiles := 0
	for w.Step() {
		if err := w.Err(); err != nil {
			return "", err
		}
		fi := w.Stat()
		if fi.Mode().IsDir() && ((fi.Name() != "." && strings.HasPrefix(fi.Name(), ".")) || fi.Name() == "cmd" || fi.Name() == "examples" || fi.Name() == "Godeps" || fi.Name() == "vendor" || fi.Name() == "third_party" || strings.HasPrefix(fi.Name(), "_") || strings.Count(w.Path(), "/") >= maxSlashes) {
			w.SkipDir()
			continue
		}
		if strings.HasSuffix(fi.Name(), ".go") {
			if numFiles >= maxFiles {
				// Instead of breaking, we SkipDir here so that we
				// ensure we always read all files in the root dir (to
				// improve the heuristic hit rate). We will not read
				// any more subdir files after calling SkipDir, which
				// is what we want.
				w.SkipDir()
			}
			numFiles++

			// For perf, read for the canonical import path 1 file at
			// a time instead of using build.Import, which always
			// reads all the files.
			contents, err := ctxvfs.ReadFile(ctx, fs, w.Path())
			if err != nil {
				return "", err
			}
			canonImportPath, err := readCanonicalImportPath(contents)
			if err == nil && canonImportPath != "" {
				// Chop off the subpackage path.
				parts := strings.Split(canonImportPath, "/")
				popComponents := strings.Count(w.Path(), "/") - 1
				if len(parts) <= popComponents {
					return "", fmt.Errorf("invalid canonical import path %q in file at path %q", canonImportPath, w.Path())
				}
				return strings.Join(parts[:len(parts)-popComponents], "/"), nil
			}
		}
	}

	// No canonical import path found, using our heuristics. Use the
	// root import path derived from the repo's clone URL.
	return rootImportPath, nil
}

func readCanonicalImportPath(contents []byte) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", contents, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return "", err
	}
	if len(f.Comments) > 0 {
		// Take the last comment (`package foo // import "xyz"`), so
		// that we avoid copyright notices and package comments.
		c := f.Comments[len(f.Comments)-1]
		txt := strings.TrimSpace(c.Text())
		if strings.HasPrefix(txt, `import "`) && strings.HasSuffix(txt, `"`) {
			return strings.TrimSuffix(strings.TrimPrefix(txt, `import "`), `"`), nil
		}
	}
	return "", nil
}

var pathHasGitSuffix = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "golangserver",
	Subsystem: "build",
	Name:      "path_has_git_suffix",
	Help:      "Temporary counter to determine if paths have a git suffix.",
})

func init() {
	prometheus.MustRegister(pathHasGitSuffix)
}
