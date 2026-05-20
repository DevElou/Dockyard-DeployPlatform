package projectservice

import (
	"context"
	"errors"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	services repository.ProjectServiceRepository
}

func NewService(services repository.ProjectServiceRepository) *Service {
	return &Service{services: services}
}

type CreateServiceInput struct {
	Name            string `json:"name"`
	ContainerPort   int    `json:"containerPort"`
	HealthcheckPath string `json:"healthcheckPath"`
	HealthcheckPort int    `json:"healthcheckPort"`
	RoutingEnabled  bool   `json:"routingEnabled"`
}

func (s *Service) List(ctx context.Context, projectID string) ([]domain.ProjectService, error) {
	return s.services.List(ctx, projectID)
}

func (s *Service) Create(ctx context.Context, projectID string, input CreateServiceInput) (domain.ProjectService, error) {
	if strings.TrimSpace(input.Name) == "" {
		return domain.ProjectService{}, errors.New("service name is required")
	}
	if input.ContainerPort < 1 || input.ContainerPort > 65535 {
		return domain.ProjectService{}, errors.New("containerPort must be between 1 and 65535")
	}

	ps := domain.ProjectService{
		ProjectID:       projectID,
		Name:            strings.TrimSpace(input.Name),
		ContainerPort:   input.ContainerPort,
		HealthcheckPath: input.HealthcheckPath,
		HealthcheckPort: input.HealthcheckPort,
		RoutingEnabled:  input.RoutingEnabled,
	}

	return s.services.Create(ctx, ps)
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.ProjectService, error) {
	return s.services.GetByID(ctx, id)
}
