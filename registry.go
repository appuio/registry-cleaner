package arc

// Repository represents a repository in a Docker registry
type Repository struct {
	Name              string
	ManifestRevisions []string
	Layers            []string
	Uploads           []string
}

// HasManifests returns true if the repository references at least one manifestRevision
func (r Repository) HasManifests() bool {
	return len(r.ManifestRevisions) > 0
}

// HasLayers returns true if the repository references at least one layer
func (r Repository) HasLayers() bool {
	return len(r.ManifestRevisions) > 0
}

// HasUploads returns true if the repository references at least one upload
func (r Repository) HasUploads() bool {
	return len(r.ManifestRevisions) > 0
}
