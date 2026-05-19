//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func setupDeploymentFixtures(t *testing.T, pool interface{}) (projectID, releaseID, targetID string) {
	t.Helper()
	ctx := context.Background()

	p, err := postgres.NewProjectRepository(setupTestDB(t)).Create(ctx, baseProject("deploy-proj"))
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return p.ID, "", ""
}

func TestDeploymentRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("d-proj-1"))
	rt, _ := postgres.NewRuntimeTargetRepository(pool).Create(ctx, baseRuntimeTarget("d-target-1"))
	rel, _ := postgres.NewReleaseRepository(pool).Create(ctx, baseRelease(p.ID, "v1.0.0"))

	repo := postgres.NewDeploymentRepository(pool)
	d := domain.Deployment{
		ProjectID:       p.ID,
		ReleaseID:       rel.ID,
		RuntimeTargetID: rt.ID,
		Strategy:        "recreate",
		Status:          domain.DeploymentStatusPending,
	}

	created, err := repo.Create(ctx, d)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}

	deployments, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(deployments) != 1 {
		t.Fatalf("expected 1 deployment, got %d", len(deployments))
	}
}

func TestDeploymentRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewDeploymentRepository(pool)
	_, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrDeploymentNotFound {
		t.Errorf("expected ErrDeploymentNotFound, got %v", err)
	}
}

func TestDeploymentRepository_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("d-proj-status"))
	rt, _ := postgres.NewRuntimeTargetRepository(pool).Create(ctx, baseRuntimeTarget("d-target-status"))
	rel, _ := postgres.NewReleaseRepository(pool).Create(ctx, baseRelease(p.ID, "v1.0.0"))

	repo := postgres.NewDeploymentRepository(pool)
	created, _ := repo.Create(ctx, domain.Deployment{
		ProjectID:       p.ID,
		ReleaseID:       rel.ID,
		RuntimeTargetID: rt.ID,
		Strategy:        "recreate",
		Status:          domain.DeploymentStatusPending,
	})

	if err := repo.UpdateStatus(ctx, created.ID, domain.DeploymentStatusDeploying, nil, nil); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	updated, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if updated.Status != domain.DeploymentStatusDeploying {
		t.Errorf("expected status %q, got %q", domain.DeploymentStatusDeploying, updated.Status)
	}
}

func TestDeploymentRepository_UpdateStatus_UnknownID(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewDeploymentRepository(pool)
	err := repo.UpdateStatus(context.Background(), "00000000-0000-0000-0000-000000000001", domain.DeploymentStatusFailed, nil, nil)
	if err != postgres.ErrDeploymentNotFound {
		t.Errorf("expected ErrDeploymentNotFound, got %v", err)
	}
}

func TestDeploymentRepository_Rollback_PreservesOriginal(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	p, _ := postgres.NewProjectRepository(pool).Create(ctx, baseProject("d-proj-rollback"))
	rt, _ := postgres.NewRuntimeTargetRepository(pool).Create(ctx, baseRuntimeTarget("d-target-rollback"))
	rel, _ := postgres.NewReleaseRepository(pool).Create(ctx, baseRelease(p.ID, "v1.0.0"))

	repo := postgres.NewDeploymentRepository(pool)
	original, _ := repo.Create(ctx, domain.Deployment{
		ProjectID:       p.ID,
		ReleaseID:       rel.ID,
		RuntimeTargetID: rt.ID,
		Strategy:        "recreate",
		Status:          domain.DeploymentStatusPending,
	})

	rollback, err := repo.Create(ctx, domain.Deployment{
		ProjectID:              p.ID,
		ReleaseID:              rel.ID,
		RuntimeTargetID:        rt.ID,
		Strategy:               "recreate",
		Status:                 domain.DeploymentStatusPending,
		RollbackOfDeploymentID: &original.ID,
	})
	if err != nil {
		t.Fatalf("create rollback: %v", err)
	}
	if rollback.RollbackOfDeploymentID == nil || *rollback.RollbackOfDeploymentID != original.ID {
		t.Error("rollback should reference original deployment ID")
	}

	// Original is unchanged
	fetched, err := repo.GetByID(ctx, original.ID)
	if err != nil {
		t.Fatalf("get original: %v", err)
	}
	if fetched.RollbackOfDeploymentID != nil {
		t.Error("original deployment should not reference a rollback")
	}
}
