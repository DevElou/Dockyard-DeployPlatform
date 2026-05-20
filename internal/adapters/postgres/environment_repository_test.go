//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func baseEnvironmentSet(projectID, name string) domain.EnvironmentSet {
	return domain.EnvironmentSet{ProjectID: projectID, Name: name}
}

func TestEnvironmentSetRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("env-proj-1"))
	repo := postgres.NewEnvironmentSetRepository(pool)

	created, err := repo.Create(ctx, baseEnvironmentSet(p.ID, "production"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	sets, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}
}

func TestEnvironmentSetRepository_DuplicateName(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("env-proj-dup"))
	repo := postgres.NewEnvironmentSetRepository(pool)

	if _, err := repo.Create(ctx, baseEnvironmentSet(p.ID, "production")); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := repo.Create(ctx, baseEnvironmentSet(p.ID, "production"))
	if err != postgres.ErrEnvironmentSetNameExists {
		t.Errorf("expected ErrEnvironmentSetNameExists, got %v", err)
	}
}

func TestEnvironmentSetRepository_GetDefaultForProject_Empty(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("env-proj-empty"))
	repo := postgres.NewEnvironmentSetRepository(pool)

	_, err := repo.GetDefaultForProject(ctx, p.ID)
	if err != postgres.ErrEnvironmentSetNotFound {
		t.Errorf("expected ErrEnvironmentSetNotFound, got %v", err)
	}
}

func TestEnvironmentVariableRepository_UpsertAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("env-proj-vars"))
	set, _ := postgres.NewEnvironmentSetRepository(pool).Create(ctx, baseEnvironmentSet(p.ID, "production"))
	repo := postgres.NewEnvironmentVariableRepository(pool)

	if err := repo.Upsert(ctx, set.ID, "DATABASE_URL", "postgres://localhost/mydb", false); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if err := repo.Upsert(ctx, set.ID, "SECRET_KEY", "supersecret", true); err != nil {
		t.Fatalf("Upsert secret: %v", err)
	}

	vars, err := repo.ListBySet(ctx, set.ID)
	if err != nil {
		t.Fatalf("ListBySet: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	for _, v := range vars {
		switch v.Key {
		case "DATABASE_URL":
			if v.Value != "postgres://localhost/mydb" {
				t.Errorf("DATABASE_URL value: got %q", v.Value)
			}
		case "SECRET_KEY":
			if !v.IsSecret {
				t.Error("expected SECRET_KEY to be secret")
			}
		}
	}
}

func TestEnvironmentVariableRepository_Upsert_UpdatesExisting(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("env-proj-upd"))
	set, _ := postgres.NewEnvironmentSetRepository(pool).Create(ctx, baseEnvironmentSet(p.ID, "production"))
	repo := postgres.NewEnvironmentVariableRepository(pool)

	if err := repo.Upsert(ctx, set.ID, "PORT", "3000", false); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := repo.Upsert(ctx, set.ID, "PORT", "8080", false); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	vars, _ := repo.ListBySet(ctx, set.ID)
	if len(vars) != 1 {
		t.Fatalf("expected 1 var after upsert, got %d", len(vars))
	}
	if vars[0].Value != "8080" {
		t.Errorf("expected updated value %q, got %q", "8080", vars[0].Value)
	}
}

func TestEnvironmentVariableRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("env-proj-del"))
	set, _ := postgres.NewEnvironmentSetRepository(pool).Create(ctx, baseEnvironmentSet(p.ID, "production"))
	repo := postgres.NewEnvironmentVariableRepository(pool)

	if err := repo.Upsert(ctx, set.ID, "TO_DELETE", "value", false); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	vars, _ := repo.ListBySet(ctx, set.ID)
	if len(vars) == 0 {
		t.Fatal("expected var to exist before delete")
	}

	if err := repo.Delete(ctx, vars[0].ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	remaining, _ := repo.ListBySet(ctx, set.ID)
	if len(remaining) != 0 {
		t.Errorf("expected 0 vars after delete, got %d", len(remaining))
	}
}

func TestEnvironmentVariableRepository_Delete_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewEnvironmentVariableRepository(pool)
	err := repo.Delete(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrEnvironmentVariableNotFound {
		t.Errorf("expected ErrEnvironmentVariableNotFound, got %v", err)
	}
}
