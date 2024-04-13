package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/sveltekit/tags"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/routevar"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/grafana/regexp"

	"path/filepath"
)

type routeInfo struct {
	Id   string
	Pattern string
	Tags []string
}

var goFileTemplate = `package sveltekit
// This file is automatically generated with gen_routes. Do not edit it directly.
import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/sveltekit/tags"
)

type svelteKitRoute struct {
	// The SvelteKit route ID
	Id      string
	// The regular expression pattern that matches the corresponding path
	Pattern *regexp.Regexp
	// The tags associated with the route
	Tag     tags.Tag
}

var svelteKitRoutes = []svelteKitRoute{
	{{range .Routes}}{
		Id:      "{{.Id}}",
		Pattern: regexp.MustCompile("{{.Pattern}}"),
		{{- if len .Tags}}
		Tag:     {{ joinTags .Tags " | " }},
		{{- end}}
	},
	{{end}}
}
`

var tsFileTemplate = `// This file is automatically generated with gen_routes. Do not edit it directly.
export type SvelteKitRoute = {
	// The SvelteKit route ID
	id: string
	// The regular expression pattern that matches the corresponding path
	pattern: RegExp
	// Whether the route is the repository root
	isRepoRoot: boolean
}

export const svelteKitRoutes: SvelteKitRoute[] = [
	{{range .Routes}}{
		id: '{{.Id}}',
		pattern: new RegExp('{{.Pattern}}'),
		isRepoRoot: {{ isRepoRoot . }},
	},
	{{end}}
]
`

const SRC_ROUTES_PREFIX = "src/routes"

func main() {
	routes := []routeInfo{}

	dest := flag.String("d", ".", "output directory")
	flag.Parse()

	for _, path := range getPagePaths() {
		i := strings.Index(path, SRC_ROUTES_PREFIX)
		routeID := filepath.Dir(path)
		if i != -1 {
			routeID = routeID[i+len(SRC_ROUTES_PREFIX):]
		}
		if routeID == "" {
			routeID = "/"
		}
		routeInfo, err := getRouteInfo(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting route info for %s: %v", path, err)
			os.Exit(1)
		}
		routeInfo.Id = routeID
		pattern, err := patternForRouteId(routeID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting pattern for route %s: %v", routeID, err)
			os.Exit(1)
		}
		routeInfo.Pattern = pattern

		routes = append(routes, routeInfo)
	}

	// Write routes to _routes.go file
	writeGoFile(filepath.Join(*dest, "_routes.go"), routes)
	writeTSFile(filepath.Join(*dest, "_routes.ts"), routes)
}

func writeGoFile(dest string, routes []routeInfo) error {
	t := template.Must(template.New("routes").Funcs(template.FuncMap{
		"joinTags": func(tags []string, sep string) string {
			if len(tags) == 0 {
				return ""
			}
			return "tags." + strings.Join(tags, " | tags.")
		},
	}).Parse(goFileTemplate))

	// Write template to _routes.go file
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	return t.Execute(file, map[string]interface{}{
		"Routes": routes,
	})
}

func writeTSFile(dest string, routes []routeInfo) error {
	t := template.Must(template.New("routes").Funcs(template.FuncMap{
		"isRepoRoot": func(route routeInfo) bool {
			for _, tag := range route.Tags {
				if tag == "RepoRoot" {
					return true
				}
			}
			return false
			},
		}).Parse(tsFileTemplate))

	// Write template to _routes.ts file
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	return t.Execute(file, map[string]interface{}{
		"Routes": routes,
	})
}

func getPagePaths() []string {
	var paths []string

	// When running under Bazel, the page file paths are passed as arguments.
	args := flag.Args()
	if len(args) > 0 {
		return args
	}

	err := filepath.Walk(SRC_ROUTES_PREFIX, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == "+page.svelte" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v", err)
		os.Exit(1)
	}
	return paths
}

var (
	splitPattern = regexp.MustCompile(`\s+`)
	tagPattern = regexp.MustCompile(`^\s*//\s+@sg\s+`)
	groupPattern = regexp.MustCompile(`^\([^)]+\)$`)
	paramPattern = regexp.MustCompile(`^(\[)?(\.\.\.)?(\w+)(?:=(\w+))?(\])?$`)
	restParamPattern = regexp.MustCompile(`^\[\.\.\.(\w+)(?:=(\w+))?\]$`)
	optionalParamPattern = regexp.MustCompile(`^\[\[(\w+)(?:=(\w+))?\]\]$`)
)

// extractTags finds lines of the forms
//     // @sg tag1 tag2 tag3
// and returns the tags found.
func getRouteInfo(path string) (routeInfo, error) {
	var info routeInfo

	f, err := os.Open(path)
	if err != nil {
		return info, err
	}
	defer f.Close()

	// All routes are opt-in by default
	tagNames := []string{"EnableOptIn"}


	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if tagPattern.MatchString(line) {
			match := tagPattern.FindString(line)
			tagNames = append(tagNames, splitPattern.Split(strings.TrimPrefix(line, match), -1)...)
			break
		}
	}

	for _, tag := range tagNames {
		if !tags.IsTagValid(tag) {
			return info, errors.Newf("Invalid tag '%s'. Valid tags: %s", tag, tags.AvailableTags())
		}
	}
	info.Tags = tagNames

	if err := scanner.Err(); err != nil {
		return info, err
	}

	return info, nil
}

// Map SvelteKit specific parameter matchers to regular expressions. This is a "best effort" approach
// because parameter matchers in SvelteKit are functions that can perform arbitrary logic.
var paramMatchers = map[string]string{
	"reporev": "/" + routevar.RepoPatternNonCapturing + `(?:@` + routevar.RevPatternNonCapturing + `)?`,
}

// This code follows the regex generation logic from
// https://github.com/sveltejs/kit/blob/main/packages/kit/src/utils/routing.js
func patternForRouteId(routeId string) (string, error) {
	if routeId == "/" {
		return "^/$", nil
	}

	b := strings.Builder{}
	b.WriteByte('^')

	for _, segment := range toSegments(routeId) {
		if segment == "" {
			continue
		}

		// [...rest]
		if restParamPattern.MatchString(segment) {
			matches := restParamPattern.FindStringSubmatch(segment)
			if (len(matches) == 3) {
				if matcher, ok := paramMatchers[matches[2]]; ok {
					b.WriteString(matcher)
					continue
				}
			}
			b.WriteString(`(?:/.*)?`)
			continue
		}

		// [[optional]]
		if optionalParamPattern.MatchString(segment) {
			b.WriteString(`(?:/[^/]+)?`)
			continue
		}

		b.WriteByte('/')
		// We don't use params within a segement, e.g.
		// foo-[bar]-[[baz]], so for simplicity we don't support that.
		if strings.Contains(segment, "[") {
			return "", errors.New("params within a segment are not supported")
		}

		b.WriteString(regexp.QuoteMeta(segment))
	}

	b.WriteString(`/?$`)

	pattern := b.String()
	if _, err := regexp.Compile(pattern); err != nil {
		return "", errors.Newf("unable to compile regexp '%q': '%v'", pattern, err)
	}

	return b.String(), nil
}

// toSegements converts a routeId to a slice of relevat segments.
// It skips empty segments and group segments.
func toSegments(routeId string) []string {
	var segments []string
	for _, segment := range strings.Split(routeId, "/") {
		if segment == "" || groupPattern.MatchString(segment) {
			continue
		}
		segments = append(segments, segment)
	}

	return segments
}
