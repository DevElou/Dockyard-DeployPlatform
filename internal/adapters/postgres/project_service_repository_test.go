//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func baseProjectService(projectID, name string, port int) domain.ProjectService {
	return domain.ProjectService{
		ProjectID:      projectID,
		Name:           name,
		ContainerPort:  port,
		RoutingEnabled: true,
	}
}

func TestProjectServiceRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("svc-proj-1"))
	repo := postgres.NewProjectServiceRepository(pool)

	created, err := repo.Create(ctx, baseProjectService(p.ID, "web", 3000))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	services, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].ContainerPort != 3000 {
		t.Errorf("port: got %d, want 3000", services[0].ContainerPort)
	}
}

func TestProjectServiceRepository_Create_DuplicateName(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("svc-proj-dup"))
	repo := postgres.NewProjectServiceRepository(pool)

	if _, err := repo.Create(ctx, baseProjectService(p.ID, "web", 3000)); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := repo.Create(ctx, baseProjectService(p.ID, "web", 8080))
	if err != postgres.ErrProjectServiceNameExists {
		t.Errorf("expected ErrProjectServiceNameExists, got %v", err)
	}
}

func TestProjectServiceRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewProjectServiceRepository(pool)
	_, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrProjectServiceNotFound {
		t.Errorf("expected ErrProjectServiceNotFound, got %v", err)
	}
}

func TestProjectServiceRepository_GetDefaultForProject(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("svc-proj-default"))
	repo := postgres.NewProjectServiceRepository(pool)

	_, err := repo.GetDefaultForProject(ctx, p.ID)
	if err != postgres.ErrProjectServiceNotFound {
		t.Errorf("expected ErrProjectServiceNotFound for empty project, got %v", err)
	}

	if _, err := repo.Create(ctx, baseProjectService(p.ID, "web", 3000)); err != nil {
		t.Fatalf("create: %v", err)
	}

	svc, err := repo.GetDefaultForProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetDefaultForProject: %v", err)
	}
	if svc.Name != "web" {
		t.Errorf("name: got %q, want %q", svc.Name, "web")
	}
}

func TestProjectServiceRepository_FieldsRoundtrip(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("svc-proj-rt"))
	repo := postgres.NewProjectServiceRepository(pool)

	input := domain.ProjectService{
		ProjectID:       p.ID,
		Name:            "api",
		ContainerPort:   8080,
		HealthcheckPath: "/healthz",
		HealthcheckPort: 8081,
		RoutingEnabled:  true,
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	checks := map[string][2]any{
		"Name":            {got.Name, "api"},
		"ContainerPort":   {got.ContainerPort, 8080},
		"HealthcheckPath": {got.HealthcheckPath, "/healthz"},
		"HealthcheckPort": {got.HealthcheckPort, 8081},
		"RoutingEnabled":  {got.RoutingEnabled, true},
	}
	for field, pair := range checks {
		if pair[0] != pair[1] {
			t.Errorf("%s: got %v, want %v", field, pair[0], pair[1])
		}
	}
}
