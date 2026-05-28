package runtimetarget

import (
	"context"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	targets repository.RuntimeTargetRepository
}

func NewService(targets repository.RuntimeTargetRepository) *Service {
	return &Service{targets: targets}
}

type CreateRuntimeTargetInput struct {
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	Endpoint     string  `json:"endpoint"`
	AgentKeyHash string  `json:"agentKeyHash"`
	AgentKey     string  `json:"agentKey"`
	ServerGroup  *string `json:"serverGroup"`
	Region       *string `json:"region"`
}

func (s *Service) List(ctx context.Context) ([]domain.RuntimeTarget, error) {
	return s.targets.List(ctx)
}

func (s *Service) Create(ctx context.Context, input CreateRuntimeTargetInput) (domain.RuntimeTarget, error) {
	agentKeyHash := strings.TrimSpace(input.AgentKeyHash)
	if agentKeyHash == "" {
		agentKeyHash = strings.TrimSpace(input.AgentKey)
	}

	rt := domain.RuntimeTarget{
		Slug:         strings.TrimSpace(input.Slug),
		Name:         strings.TrimSpace(input.Name),
		RuntimeType:  domain.RuntimeTypeDocker,
		Endpoint:     strings.TrimSpace(input.Endpoint),
		AgentKeyHash: agentKeyHash,
		ServerGroup:  input.ServerGroup,
		Region:       input.Region,
		Enabled:      true,
	}

	if err := rt.Validate(); err != nil {
		return domain.RuntimeTarget{}, err
	}

	return s.targets.Create(ctx, rt)
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.RuntimeTarget, error) {
	return s.targets.GetByID(ctx, id)
}

func (s *Service) Enable(ctx context.Context, id string) error {
	return s.targets.SetEnabled(ctx, id, true)
}

func (s *Service) Disable(ctx context.Context, id string) error {
	return s.targets.SetEnabled(ctx, id, false)
}
