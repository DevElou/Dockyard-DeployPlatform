package release

import (
	"context"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/source"
)

type Service struct {
	releases repository.ReleaseRepository
	source   source.Provider
}

func NewService(releases repository.ReleaseRepository, src source.Provider) *Service {
	return &Service{releases: releases, source: src}
}

type CreateReleaseInput struct {
	Version         string  `json:"version"`
	GitRef          string  `json:"gitRef"`
	CreatedByUserID *string `json:"createdByUserId"`
}

func (s *Service) List(ctx context.Context, projectID string) ([]domain.Release, error) {
	return s.releases.List(ctx, projectID)
}

func (s *Service) Create(ctx context.Context, projectID string, input CreateReleaseInput) (domain.Release, error) {
	rev, err := s.source.ResolveRevision(ctx, projectID, strings.TrimSpace(input.GitRef))
	if err != nil {
		return domain.Release{}, err
	}

	release := domain.Release{
		ProjectID:       strings.TrimSpace(projectID),
		Version:         strings.TrimSpace(input.Version),
		SourceType:      "github",
		GitSHA:          rev.CommitSHA,
		GitRef:          rev.GitRef,
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
