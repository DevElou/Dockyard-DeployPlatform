package deployment

import (
	"context"
	"strings"
	"time"

	"github.com/elouan/dockyard/internal/application/operationlog"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	deployments repository.DeploymentRepository
	events      *operationlog.Service
}

func NewService(deployments repository.DeploymentRepository, events *operationlog.Service) *Service {
	return &Service{deployments: deployments, events: events}
}

type CreateDeploymentInput struct {
	ReleaseID       string  `json:"releaseId"`
	RuntimeTargetID string  `json:"runtimeTargetId"`
	Strategy        string  `json:"strategy"`
	TriggeredByUser *string `json:"triggeredByUserId"`
}

type UpdateStatusInput struct {
	Status     domain.DeploymentStatus `json:"status"`
	StartedAt  *time.Time              `json:"startedAt"`
	FinishedAt *time.Time              `json:"finishedAt"`
}

func (s *Service) List(ctx context.Context, projectID string) ([]domain.Deployment, error) {
	return s.deployments.List(ctx, projectID)
}

func (s *Service) Create(ctx context.Context, projectID string, input CreateDeploymentInput) (domain.Deployment, error) {
	d := domain.Deployment{
		ProjectID:         strings.TrimSpace(projectID),
		ReleaseID:         strings.TrimSpace(input.ReleaseID),
		RuntimeTargetID:   strings.TrimSpace(input.RuntimeTargetID),
		Strategy:          defaultString(input.Strategy, "recreate"),
		Status:            domain.DeploymentStatusPending,
		TriggeredByUserID: input.TriggeredByUser,
	}

	if err := d.Validate(); err != nil {
		return domain.Deployment{}, err
	}

	created, err := s.deployments.Create(ctx, d)
	if err != nil {
		return domain.Deployment{}, err
	}

	if s.events != nil {
		s.events.Info(ctx, domain.OperationResourceDeployment, created.ID, "queued",
			"deployment created and waiting for worker",
			map[string]string{
				"releaseId":       created.ReleaseID,
				"runtimeTargetId": created.RuntimeTargetID,
				"strategy":        created.Strategy,
			})
	}
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Deployment, error) {
	return s.deployments.GetByID(ctx, id)
}

// Rollback creates a new deployment pointing to the same release as the original.
func (s *Service) Rollback(ctx context.Context, projectID, originalDeploymentID string) (domain.Deployment, error) {
	original, err := s.deployments.GetByID(ctx, originalDeploymentID)
	if err != nil {
		return domain.Deployment{}, err
	}

	rollback := domain.Deployment{
		ProjectID:              projectID,
		ReleaseID:              original.ReleaseID,
		RuntimeTargetID:        original.RuntimeTargetID,
		Strategy:               original.Strategy,
		Status:                 domain.DeploymentStatusPending,
		RollbackOfDeploymentID: &originalDeploymentID,
	}

	if err := rollback.Validate(); err != nil {
		return domain.Deployment{}, err
	}

	return s.deployments.Create(ctx, rollback)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, input UpdateStatusInput) error {
	return s.deployments.UpdateStatus(ctx, id, input.Status, input.StartedAt, input.FinishedAt)
}

func (s *Service) ListEvents(ctx context.Context, deploymentID string) ([]domain.OperationEvent, error) {
	if s.events == nil {
		return []domain.OperationEvent{}, nil
	}
	return s.events.ListForResource(ctx, domain.OperationResourceDeployment, deploymentID)
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
