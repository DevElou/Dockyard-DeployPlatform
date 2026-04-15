package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type ProjectRepository interface {
	List(ctx context.Context) ([]domain.Project, error)
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
}
