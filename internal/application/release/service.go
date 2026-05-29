package release

import (
	"context"
	"strings"

	"github.com/elouan/dockyard/internal/application/operationlog"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/source"
)

type Service struct {
	releases repository.ReleaseRepository
	source   source.Provider
	events   *operationlog.Service
}

func NewService(releases repository.ReleaseRepository, src source.Provider, events *operationlog.Service) *Service {
	return &Service{releases: releases, source: src, events: events}
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

	created, err := s.releases.Create(ctx, release)
	if err != nil {
		return domain.Release{}, err
	}

	if s.events != nil {
		s.events.Info(ctx, domain.OperationResourceRelease, created.ID, "resolving_source",
			"resolved git ref to commit",
			map[string]string{"gitRef": rev.GitRef, "gitSha": rev.CommitSHA})
	}
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Release, error) {
	return s.releases.GetByID(ctx, id)
}

func (s *Service) ListEvents(ctx context.Context, releaseID string) ([]domain.OperationEvent, error) {
	if s.events == nil {
		return []domain.OperationEvent{}, nil
	}
	return s.events.ListForResource(ctx, domain.OperationResourceRelease, releaseID)
}
