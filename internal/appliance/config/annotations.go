package config

const (
	ConfigmapName = "sourcegraph-appliance"

	AnnotationKeyManaged        = "appliance.sourcegraph.com/managed"
	AnnotationKeyCurrentVersion = "appliance.sourcegraph.com/currentVersion"
	AnnotationKeyConfigHash     = "appliance.sourcegraph.com/configHash"
)
