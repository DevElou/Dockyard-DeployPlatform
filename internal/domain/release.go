package domain

import (
	"errors"
	"strings"
	"time"
)

type Release struct {
	ID              string      `json:"id"`
	ProjectID       string      `json:"projectId"`
	Version         string      `json:"version"`
	SourceType      string      `json:"sourceType"`
	GitSHA          string      `json:"gitSha"`
	GitRef          string      `json:"gitRef"`
	ImageRepository string      `json:"imageRepository"`
	ImageTag        string      `json:"imageTag"`
	ImageDigest     string      `json:"imageDigest"`
	BuildStatus     BuildStatus `json:"buildStatus"`
	CreatedByUserID *string     `json:"createdByUserId"`
	CreatedAt       time.Time   `json:"createdAt"`
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
	// image fields are populated async by BuildWorker
	return nil
}
