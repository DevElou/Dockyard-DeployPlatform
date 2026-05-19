package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type ReleaseRepository interface {
	List(ctx context.Context, projectID string) ([]domain.Release, error)
	Create(ctx context.Context, release domain.Release) (domain.Release, error)
	GetByID(ctx context.Context, id string) (domain.Release, error)
}
