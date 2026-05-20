package deployment

import (
	"context"
	"strings"
	"time"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	deployments repository.DeploymentRepository
}

func NewService(deployments repository.DeploymentRepository) *Service {
	return &Service{deployments: deployments}
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
		ProjectID:       strings.TrimSpace(projectID),
		ReleaseID:       strings.TrimSpace(input.ReleaseID),
		RuntimeTargetID: strings.TrimSpace(input.RuntimeTargetID),
		Strategy:        defaultString(input.Strategy, "recreate"),
		Status:          domain.DeploymentStatusPending,
		TriggeredByUserID: input.TriggeredByUser,
	}

	if err := d.Validate(); err != nil {
		return domain.Deployment{}, err
	}

	return s.deployments.Create(ctx, d)
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

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
