package release

import (
	"context"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	releases repository.ReleaseRepository
}

func NewService(releases repository.ReleaseRepository) *Service {
	return &Service{releases: releases}
}

type CreateReleaseInput struct {
	Version         string  `json:"version"`
	GitSHA          string  `json:"gitSha"`
	GitRef          string  `json:"gitRef"`
	ImageRepository string  `json:"imageRepository"`
	ImageTag        string  `json:"imageTag"`
	ImageDigest     string  `json:"imageDigest"`
	CreatedByUserID *string `json:"createdByUserId"`
}

func (s *Service) List(ctx context.Context, projectID string) ([]domain.Release, error) {
	return s.releases.List(ctx, projectID)
}

func (s *Service) Create(ctx context.Context, projectID string, input CreateReleaseInput) (domain.Release, error) {
	release := domain.Release{
		ProjectID:       strings.TrimSpace(projectID),
		Version:         strings.TrimSpace(input.Version),
		SourceType:      "github",
		GitSHA:          strings.TrimSpace(input.GitSHA),
		GitRef:          strings.TrimSpace(input.GitRef),
		ImageRepository: strings.TrimSpace(input.ImageRepository),
		ImageTag:        strings.TrimSpace(input.ImageTag),
		ImageDigest:     strings.TrimSpace(input.ImageDigest),
		BuildStatus:     domain.BuildStatusPending,
		CreatedByUserID: input.CreatedByUserID,
	}

	if err := release.Validate(); err != nil {
		return domain.Release{}, err
	}

	return s.releases.Create(ctx, release)
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Release, error) {
	return s.releases.GetByID(ctx, id)
}
