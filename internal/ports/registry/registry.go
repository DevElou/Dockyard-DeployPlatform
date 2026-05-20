package registry

import "context"

type BuildRequest struct {
	ProjectID      string
	ReleaseVersion string
	CommitSHA      string
	BuildContext   string
	DockerfilePath string
}

type BuildResult struct {
	ImageRepository string
	ImageTag        string
	ImageDigest     string
}

type Builder interface {
	BuildAndPush(ctx context.Context, request BuildRequest) (BuildResult, error)
}
