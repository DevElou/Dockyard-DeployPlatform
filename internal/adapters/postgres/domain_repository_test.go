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
