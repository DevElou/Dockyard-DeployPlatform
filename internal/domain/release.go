package domain

type Release struct {
	ID              string
	ProjectID       string
	Version         string
	GitSHA          string
	GitRef          string
	ImageRepository string
	ImageTag        string
	ImageDigest     string
	BuildStatus     BuildStatus
	CreatedAt       string
}
