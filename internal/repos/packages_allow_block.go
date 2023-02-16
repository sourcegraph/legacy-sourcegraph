package repos

import (
	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type PackageMatcher interface {
	Matches(pkg, version string) bool
}

type packageNameGlob struct {
	g glob.Glob
}

func NewPackageNameGlob(str string) (PackageMatcher, error) {
	g, err := glob.Compile(str)
	if err != nil {
		return nil, err
	}
	return packageNameGlob{g}, nil
}

func (p packageNameGlob) Matches(pkg, _ string) bool {
	// when the package name is to be glob matched, we dont
	// care about the version
	return p.g.Match(pkg)
}

type versionGlob struct {
	packageName, globStr string
	g                    glob.Glob
}

func NewVersionGlob(packageName, str string) (PackageMatcher, error) {
	g, err := glob.Compile(str)
	if err != nil {
		return nil, err
	}
	return versionGlob{packageName, str, g}, nil
}

func (v versionGlob) Matches(pkg, version string) bool {
	// when the version is to be glob matched, the package name
	// has to match exactly
	return pkg == v.packageName && v.g.Match(version)
}

func IsPackageAllowed(pkgName reposource.PackageName, allowList, blockList []PackageMatcher) bool {
	// blocklist takes priority
	for _, block := range blockList {
		// non-all-encompassing version globs don't apply to unversioned packages,
		// likely we're at too-early point in the syncing process to know, but also
		// we may still want the package to browse versions that _dont_ match this
		if vglob, ok := block.(versionGlob); ok && vglob.globStr != "*" {
			continue
		}

		if block.Matches(string(pkgName), "") {
			return false
		}
	}

	// by default, anything is allowed unless specific allowlist exists
	isAllowed := len(allowList) == 0
	for _, allow := range allowList {
		if vglob, ok := allow.(versionGlob); ok && vglob.globStr != "*" {
			continue
		}
		isAllowed = isAllowed || allow.Matches(string(pkgName), "")
	}

	return isAllowed
}

func IsVersionedPackageAllowed(pkgName reposource.PackageName, version string, allowList, blockList []PackageMatcher) bool {
	// blocklist takes priority
	for _, block := range blockList {
		if _, ok := block.(versionGlob); ok && version == "" {
			continue
		}

		if block.Matches(string(pkgName), version) {
			return false
		}
	}

	// by default, anything is allowed unless specific allowlist exists
	isAllowed := len(allowList) == 0
	for _, allow := range allowList {
		if _, ok := allow.(versionGlob); ok && version == "" {
			continue
		}
		isAllowed = isAllowed || allow.Matches(string(pkgName), version)
	}

	return isAllowed
}
