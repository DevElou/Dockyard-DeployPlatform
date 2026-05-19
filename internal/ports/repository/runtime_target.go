package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type RuntimeTargetRepository interface {
	List(ctx context.Context) ([]domain.RuntimeTarget, error)
	Create(ctx context.Context, rt domain.RuntimeTarget) (domain.RuntimeTarget, error)
	GetByID(ctx context.Context, id string) (domain.RuntimeTarget, error)
	SetEnabled(ctx context.Context, id string, enabled bool) error
}
