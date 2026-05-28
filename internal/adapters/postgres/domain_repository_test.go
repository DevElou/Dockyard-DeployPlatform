//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func baseDomain(projectID, hostname string) domain.Domain {
	return domain.Domain{
		ProjectID:   projectID,
		Hostname:    hostname,
		BaseDomain:  "duckdns.org",
		Provider:    "duckdns",
		RoutingType: "host",
		TLSEnabled:  true,
		Status:      domain.DomainStatusPending,
	}
}

func TestDomainRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("domain-proj-1"))

	repo := postgres.NewDomainRepository(pool)
	created, err := repo.Create(ctx, baseDomain(p.ID, "app.duckdns.org"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	domains, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}
}

func TestDomainRepository_DuplicateHostname(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("domain-proj-dup"))

	repo := postgres.NewDomainRepository(pool)
	if _, err := repo.Create(ctx, baseDomain(p.ID, "dup.duckdns.org")); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := repo.Create(ctx, baseDomain(p.ID, "dup.duckdns.org"))
	if err != postgres.ErrDomainHostnameExists {
		t.Errorf("expected ErrDomainHostnameExists, got %v", err)
	}
}

func TestDomainRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewDomainRepository(pool)
	_, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrDomainNotFound {
		t.Errorf("expected ErrDomainNotFound, got %v", err)
	}
}

func TestDomainRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("domain-proj-del"))

	repo := postgres.NewDomainRepository(pool)
	created, _ := repo.Create(ctx, baseDomain(p.ID, "del.duckdns.org"))

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	domains, _ := repo.List(ctx, p.ID)
	if len(domains) != 0 {
		t.Errorf("expected 0 domains after delete, got %d", len(domains))
	}
}

func TestDomainRepository_Delete_UnknownID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewDomainRepository(pool)
	err := repo.Delete(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrDomainNotFound {
		t.Errorf("expected ErrDomainNotFound, got %v", err)
	}
}

func TestDomainRepository_ListByProjectService(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("domain-proj-svc"))

	svcRepo := postgres.NewProjectServiceRepository(pool)
	svc, err := svcRepo.Create(ctx, domain.ProjectService{
		ProjectID:     p.ID,
		Name:          "web",
		ContainerPort: 3000,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	domainRepo := postgres.NewDomainRepository(pool)

	// domain linked to the service
	d := baseDomain(p.ID, "svc.duckdns.org")
	d.ProjectServiceID = &svc.ID
	created, err := domainRepo.Create(ctx, d)
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}

	// domain NOT linked to any service
	_, _ = domainRepo.Create(ctx, baseDomain(p.ID, "other.duckdns.org"))

	results, err := domainRepo.ListByProjectService(ctx, svc.ID)
	if err != nil {
		t.Fatalf("ListByProjectService: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(results))
	}
	if results[0].ID != created.ID {
		t.Errorf("unexpected domain ID %s", results[0].ID)
	}

	// empty result for unknown service ID
	empty, err := domainRepo.ListByProjectService(ctx, "00000000-0000-0000-0000-000000000002")
	if err != nil {
		t.Fatalf("ListByProjectService empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 domains for unknown service, got %d", len(empty))
	}
}

func TestDomainRepository_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("domain-proj-status"))

	repo := postgres.NewDomainRepository(pool)
	created, _ := repo.Create(ctx, baseDomain(p.ID, "status.duckdns.org"))

	if err := repo.UpdateStatus(ctx, created.ID, domain.DomainStatusReady); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID after UpdateStatus: %v", err)
	}
	if got.Status != domain.DomainStatusReady {
		t.Errorf("status = %q, want ready", got.Status)
	}
}

func TestDomainRepository_UpdateStatus_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	err := postgres.NewDomainRepository(pool).UpdateStatus(
		context.Background(),
		"00000000-0000-0000-0000-000000000001",
		domain.DomainStatusReady,
	)
	if err != postgres.ErrDomainNotFound {
		t.Errorf("expected ErrDomainNotFound, got %v", err)
	}
}
