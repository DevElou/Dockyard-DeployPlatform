package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type ProjectRepository interface {
	List(ctx context.Context) ([]domain.Project, error)
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	GetByID(ctx context.Context, id string) (domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (domain.Project, error)
	Archive(ctx context.Context, id string) error
	ListRuntimeTargets(ctx context.Context, projectID string) ([]domain.RuntimeTarget, error)
	AddRuntimeTarget(ctx context.Context, projectID, runtimeTargetID string) error
}
