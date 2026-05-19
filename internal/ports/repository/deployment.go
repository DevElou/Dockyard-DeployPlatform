package repository

import (
	"context"
	"time"

	"github.com/elouan/dockyard/internal/domain"
)

type DeploymentRepository interface {
	List(ctx context.Context, projectID string) ([]domain.Deployment, error)
	Create(ctx context.Context, deployment domain.Deployment) (domain.Deployment, error)
	GetByID(ctx context.Context, id string) (domain.Deployment, error)
	UpdateStatus(ctx context.Context, id string, status domain.DeploymentStatus, startedAt *time.Time, finishedAt *time.Time) error
}
