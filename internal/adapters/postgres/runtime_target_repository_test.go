//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func baseRuntimeTarget(slug string) domain.RuntimeTarget {
	return domain.RuntimeTarget{
		Slug:         slug,
		Name:         "Test Target",
		RuntimeType:  domain.RuntimeTypeDocker,
		Endpoint:     "http://192.168.1.10:2375",
		AgentKeyHash: "abc123hash",
		Enabled:      true,
	}
}

func TestRuntimeTargetRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewRuntimeTargetRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, baseRuntimeTarget("target-1"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	targets, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", targets[0].ID, created.ID)
	}
}

func TestRuntimeTargetRepository_DuplicateSlug(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewRuntimeTargetRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, baseRuntimeTarget("dup")); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := repo.Create(ctx, baseRuntimeTarget("dup"))
	if err != postgres.ErrRuntimeTargetSlugExists {
		t.Errorf("expected ErrRuntimeTargetSlugExists, got %v", err)
	}
}

func TestRuntimeTargetRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewRuntimeTargetRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrRuntimeTargetNotFound {
		t.Errorf("expected ErrRuntimeTargetNotFound, got %v", err)
	}
}

func TestRuntimeTargetRepository_SetEnabled_TogglesFlag(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewRuntimeTargetRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, baseRuntimeTarget("toggle-target"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.SetEnabled(ctx, created.ID, false); err != nil {
		t.Fatalf("SetEnabled(false): %v", err)
	}

	target, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if target.Enabled {
		t.Error("expected Enabled=false after disable")
	}

	if err := repo.SetEnabled(ctx, created.ID, true); err != nil {
		t.Fatalf("SetEnabled(true): %v", err)
	}

	target, err = repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !target.Enabled {
		t.Error("expected Enabled=true after re-enable")
	}
}

func TestRuntimeTargetRepository_SetEnabled_UnknownID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewRuntimeTargetRepository(pool)
	ctx := context.Background()

	err := repo.SetEnabled(ctx, "00000000-0000-0000-0000-000000000001", false)
	if err != postgres.ErrRuntimeTargetNotFound {
		t.Errorf("expected ErrRuntimeTargetNotFound, got %v", err)
	}
}
