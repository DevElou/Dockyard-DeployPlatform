package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type ProjectServiceRepository interface {
	List(ctx context.Context, projectID string) ([]domain.ProjectService, error)
	Create(ctx context.Context, ps domain.ProjectService) (domain.ProjectService, error)
	GetByID(ctx context.Context, id string) (domain.ProjectService, error)
	GetDefaultForProject(ctx context.Context, projectID string) (domain.ProjectService, error)
}
