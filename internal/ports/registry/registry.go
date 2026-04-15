package registry

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
	BuildAndPush(request BuildRequest) (BuildResult, error)
}
