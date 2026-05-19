package domainsvc

import (
	"context"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	domains repository.DomainRepository
}

func NewService(domains repository.DomainRepository) *Service {
	return &Service{domains: domains}
}

type CreateDomainInput struct {
	Hostname         string  `json:"hostname"`
	BaseDomain       string  `json:"baseDomain"`
	ProjectServiceID *string `json:"projectServiceId"`
	Provider         string  `json:"provider"`
	RoutingType      string  `json:"routingType"`
	TLSEnabled       *bool   `json:"tlsEnabled"`
}

func (s *Service) List(ctx context.Context, projectID string) ([]domain.Domain, error) {
	return s.domains.List(ctx, projectID)
}

func (s *Service) Create(ctx context.Context, projectID string, input CreateDomainInput) (domain.Domain, error) {
	tlsEnabled := true
	if input.TLSEnabled != nil {
		tlsEnabled = *input.TLSEnabled
	}

	d := domain.Domain{
		ProjectID:        strings.TrimSpace(projectID),
		ProjectServiceID: input.ProjectServiceID,
		Hostname:         strings.TrimSpace(input.Hostname),
		BaseDomain:       strings.TrimSpace(input.BaseDomain),
		Provider:         defaultString(input.Provider, "duckdns"),
		RoutingType:      defaultString(input.RoutingType, "host"),
		TLSEnabled:       tlsEnabled,
		Status:           domain.DomainStatusPending,
	}

	if err := d.Validate(); err != nil {
		return domain.Domain{}, err
	}

	return s.domains.Create(ctx, d)
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Domain, error) {
	return s.domains.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.domains.Delete(ctx, id)
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
