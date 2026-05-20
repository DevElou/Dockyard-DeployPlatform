package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type ReleaseRepository interface {
	List(ctx context.Context, projectID string) ([]domain.Release, error)
	ListByBuildStatus(ctx context.Context, status domain.BuildStatus) ([]domain.Release, error)
	Create(ctx context.Context, release domain.Release) (domain.Release, error)
	GetByID(ctx context.Context, id string) (domain.Release, error)
	UpdateBuildStatus(ctx context.Context, id string, status domain.BuildStatus) error
	UpdateBuildResult(ctx context.Context, id string, imageRepository, imageTag, imageDigest string, status domain.BuildStatus) error
}
