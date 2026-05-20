package source

import "context"

type Revision struct {
	CommitSHA  string
	GitRef     string
	ArchiveURL string
}

type Provider interface {
	ResolveRevision(ctx context.Context, projectID string, ref string) (Revision, error)
	DownloadArchive(ctx context.Context, projectID string, commitSHA string, targetDir string) error
}
