package environment

import (
	"context"
	"errors"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	sets repository.EnvironmentSetRepository
	vars repository.EnvironmentVariableRepository
}

func NewService(sets repository.EnvironmentSetRepository, vars repository.EnvironmentVariableRepository) *Service {
	return &Service{sets: sets, vars: vars}
}

type CreateSetInput struct {
	Name string `json:"name"`
}

type UpsertVariableInput struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

func (s *Service) ListSets(ctx context.Context, projectID string) ([]domain.EnvironmentSet, error) {
	return s.sets.List(ctx, projectID)
}

func (s *Service) CreateSet(ctx context.Context, projectID string, input CreateSetInput) (domain.EnvironmentSet, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.EnvironmentSet{}, errors.New("environment set name is required")
	}
	return s.sets.Create(ctx, domain.EnvironmentSet{
		ProjectID: strings.TrimSpace(projectID),
		Name:      name,
	})
}

func (s *Service) ListVariables(ctx context.Context, setID string) ([]domain.EnvironmentVariable, error) {
	return s.vars.ListBySet(ctx, setID)
}

func (s *Service) UpsertVariable(ctx context.Context, setID string, input UpsertVariableInput) error {
	key := strings.TrimSpace(input.Key)
	if key == "" {
		return errors.New("variable key is required")
	}
	return s.vars.Upsert(ctx, setID, key, input.Value, input.IsSecret)
}

func (s *Service) DeleteVariable(ctx context.Context, varID string) error {
	return s.vars.Delete(ctx, varID)
}
