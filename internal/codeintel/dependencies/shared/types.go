package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/packagerepos"
)

type PackageRepoReference struct {
	ID       int
	Scheme   string
	Name     reposource.PackageName
	Versions []PackageRepoRefVersion
}

type PackageRepoRefVersion struct {
	ID           int
	PackageRefID int
	Version      string
}

type MinimalPackageRepoRef struct {
	Scheme   string
	Name     reposource.PackageName
	Versions []string
}

type MinimialVersionedPackageRepo struct {
	Scheme  string
	Name    reposource.PackageName
	Version string
}

type PackageFilter struct {
	ID              int
	Behaviour       string
	ExternalService string
	NameFilter      *struct {
		PackageGlob string
	}
	VersionFilter *struct {
		PackageName string
		VersionGlob string
	}
}

func (f *PackageFilter) BuildMatcher() (packagerepos.PackageMatcher, error) {
	if f.NameFilter != nil {
		return packagerepos.NewPackageNameGlob(f.NameFilter.PackageGlob)
	}
	return packagerepos.NewVersionGlob(f.VersionFilter.PackageName, f.VersionFilter.VersionGlob)
}
