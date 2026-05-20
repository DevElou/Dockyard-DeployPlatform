package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type EnvironmentSetRepository interface {
	List(ctx context.Context, projectID string) ([]domain.EnvironmentSet, error)
	Create(ctx context.Context, set domain.EnvironmentSet) (domain.EnvironmentSet, error)
	GetByID(ctx context.Context, id string) (domain.EnvironmentSet, error)
	GetDefaultForProject(ctx context.Context, projectID string) (domain.EnvironmentSet, error)
}

type EnvironmentVariableRepository interface {
	ListBySet(ctx context.Context, environmentSetID string) ([]domain.EnvironmentVariable, error)
	Upsert(ctx context.Context, environmentSetID, key, value string, isSecret bool) error
	Delete(ctx context.Context, id string) error
}
