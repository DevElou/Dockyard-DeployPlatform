package domainsvc

import (
	"context"
	"log"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/routing"
)

type Service struct {
	domains repository.DomainRepository
	routing routing.Provider
}

func NewService(domains repository.DomainRepository, r routing.Provider) *Service {
	return &Service{domains: domains, routing: r}
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
	d, err := s.domains.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.domains.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.routing.DeleteRoute(d.Hostname); err != nil {
		log.Printf("domainsvc: delete route %s: %v", d.Hostname, err)
	}
	return nil
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
