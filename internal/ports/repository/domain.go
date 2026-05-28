package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

type DomainRepository interface {
	List(ctx context.Context, projectID string) ([]domain.Domain, error)
	Create(ctx context.Context, d domain.Domain) (domain.Domain, error)
	GetByID(ctx context.Context, id string) (domain.Domain, error)
	Delete(ctx context.Context, id string) error
	ListByProjectService(ctx context.Context, projectServiceID string) ([]domain.Domain, error)
	UpdateStatus(ctx context.Context, id string, status domain.DomainStatus) error
}
