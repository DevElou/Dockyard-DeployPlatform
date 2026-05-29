// Package operationlog provides a thin recording API over the OperationLogRepository.
//
// Workers and HTTP handlers should depend on this service rather than the
// repository directly, so the underlying storage and retention policy can
// evolve without touching business logic.
package operationlog

import (
	"context"
	"log"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

// Recorder is the small interface workers depend on for emitting events.
type Recorder interface {
	Record(ctx context.Context, ev domain.OperationEvent) error
	Info(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string)
	Warn(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string)
	Error(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string)
	Success(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string)
}

type Service struct {
	repo repository.OperationLogRepository
}

func NewService(repo repository.OperationLogRepository) *Service {
	return &Service{repo: repo}
}

// Record appends an event verbatim. Storage failures are logged but do not
// bubble up to callers — observability must never block the workflow it
// observes.
func (s *Service) Record(ctx context.Context, ev domain.OperationEvent) error {
	if _, err := s.repo.Append(ctx, ev); err != nil {
		log.Printf("operationlog: append %s/%s/%s: %v", ev.ResourceType, ev.ResourceID, ev.Phase, err)
		return err
	}
	return nil
}

func (s *Service) Info(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string) {
	_ = s.Record(ctx, build(resource, id, phase, domain.OperationLevelInfo, msg, details))
}

func (s *Service) Warn(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string) {
	_ = s.Record(ctx, build(resource, id, phase, domain.OperationLevelWarn, msg, details))
}

func (s *Service) Error(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string) {
	_ = s.Record(ctx, build(resource, id, phase, domain.OperationLevelError, msg, details))
}

func (s *Service) Success(ctx context.Context, resource domain.OperationResourceType, id, phase, msg string, details map[string]string) {
	_ = s.Record(ctx, build(resource, id, phase, domain.OperationLevelSuccess, msg, details))
}

func (s *Service) ListForResource(ctx context.Context, resource domain.OperationResourceType, id string) ([]domain.OperationEvent, error) {
	return s.repo.List(ctx, resource, id)
}

func build(resource domain.OperationResourceType, id, phase string, level domain.OperationLevel, msg string, details map[string]string) domain.OperationEvent {
	return domain.OperationEvent{
		ResourceType: resource,
		ResourceID:   id,
		Phase:        phase,
		Level:        level,
		Message:      msg,
		Details:      details,
	}
}
