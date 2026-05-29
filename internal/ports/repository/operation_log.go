package repository

import (
	"context"

	"github.com/elouan/dockyard/internal/domain"
)

// OperationLogRepository persists OperationEvents for releases and deployments.
//
// Implementations are expected to enforce a per-resource retention bound after
// each append (default: keep the latest N events per (resourceType, resourceID)).
type OperationLogRepository interface {
	Append(ctx context.Context, event domain.OperationEvent) (domain.OperationEvent, error)
	List(ctx context.Context, resourceType domain.OperationResourceType, resourceID string) ([]domain.OperationEvent, error)
	Prune(ctx context.Context, resourceType domain.OperationResourceType, resourceID string, keep int) error
}
