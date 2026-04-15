package source

type Revision struct {
	CommitSHA  string
	GitRef     string
	ArchiveURL string
}

type Provider interface {
	ResolveRevision(projectID string, ref string) (Revision, error)
}
