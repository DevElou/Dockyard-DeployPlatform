package domain

import (
	"errors"
	"strings"
	"time"
)

type Release struct {
	ID              string
	ProjectID       string
	Version         string
	SourceType      string
	GitSHA          string
	GitRef          string
	ImageRepository string
	ImageTag        string
	ImageDigest     string
	BuildStatus     BuildStatus
	CreatedByUserID *string
	CreatedAt       time.Time
}

func (r Release) Validate() error {
	if strings.TrimSpace(r.ProjectID) == "" {
		return errors.New("release project ID is required")
	}
	if strings.TrimSpace(r.Version) == "" {
		return errors.New("release version is required")
	}
	if strings.TrimSpace(r.GitSHA) == "" {
		return errors.New("release git SHA is required")
	}
	if strings.TrimSpace(r.ImageRepository) == "" {
		return errors.New("release image repository is required")
	}
	if strings.TrimSpace(r.ImageTag) == "" {
		return errors.New("release image tag is required")
	}
	if strings.TrimSpace(r.ImageDigest) == "" {
		return errors.New("release image digest is required")
	}
	return nil
}
