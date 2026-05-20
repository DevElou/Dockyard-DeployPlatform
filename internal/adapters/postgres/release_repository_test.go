//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func insertProject(t *testing.T, pool interface{ Exec(context.Context, string, ...any) (interface{ RowsAffected() int64 }, error) }, slug string) string {
	t.Helper()
	return ""
}

func createTestProject(t *testing.T, ctx context.Context, slug string) (string, *postgres.ProjectRepository) {
	t.Helper()
	pool := setupTestDB(t)
	repo := postgres.NewProjectRepository(pool)
	p, err := repo.Create(ctx, baseProject(slug))
	if err != nil {
		t.Fatalf("create test project: %v", err)
	}
	return p.ID, repo
}

func baseRelease(projectID, version string) domain.Release {
	return domain.Release{
		ProjectID:       projectID,
		Version:         version,
		SourceType:      "github",
		GitSHA:          "abc123def456",
		GitRef:          "refs/heads/main",
		ImageRepository: "registry.local/my-app",
		ImageTag:        version,
		ImageDigest:     "sha256:" + version,
		BuildStatus:     domain.BuildStatusSucceeded,
	}
}

func TestReleaseRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p, err := projectRepo.Create(ctx, baseProject("release-proj"))
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	releaseRepo := postgres.NewReleaseRepository(pool)
	created, err := releaseRepo.Create(ctx, baseRelease(p.ID, "v1.0.0"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}

	releases, err := releaseRepo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
}

func TestReleaseRepository_DuplicateVersion(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p, _ := projectRepo.Create(ctx, baseProject("dup-version-proj"))

	releaseRepo := postgres.NewReleaseRepository(pool)
	if _, err := releaseRepo.Create(ctx, baseRelease(p.ID, "v1.0.0")); err != nil {
		t.Fatalf("first: %v", err)
	}

	r2 := baseRelease(p.ID, "v1.0.0")
	r2.ImageDigest = "sha256:different"
	_, err := releaseRepo.Create(ctx, r2)
	if err != postgres.ErrReleaseVersionExists {
		t.Errorf("expected ErrReleaseVersionExists, got %v", err)
	}
}

func TestReleaseRepository_DuplicateDigest(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p, _ := projectRepo.Create(ctx, baseProject("dup-digest-proj"))

	releaseRepo := postgres.NewReleaseRepository(pool)
	if _, err := releaseRepo.Create(ctx, baseRelease(p.ID, "v1.0.0")); err != nil {
		t.Fatalf("first: %v", err)
	}

	r2 := baseRelease(p.ID, "v2.0.0")
	r2.ImageDigest = "sha256:v1.0.0"
	_, err := releaseRepo.Create(ctx, r2)
	if err != postgres.ErrReleaseDigestExists {
		t.Errorf("expected ErrReleaseDigestExists, got %v", err)
	}
}

func TestReleaseRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewReleaseRepository(pool)
	_, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != postgres.ErrReleaseNotFound {
		t.Errorf("expected ErrReleaseNotFound, got %v", err)
	}
}

func TestReleaseRepository_List_FilterByProject(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p1, _ := projectRepo.Create(ctx, baseProject("proj-filter-1"))
	p2, _ := projectRepo.Create(ctx, baseProject("proj-filter-2"))

	releaseRepo := postgres.NewReleaseRepository(pool)
	if _, err := releaseRepo.Create(ctx, baseRelease(p1.ID, "v1.0.0")); err != nil {
		t.Fatalf("create for p1: %v", err)
	}
	if _, err := releaseRepo.Create(ctx, baseRelease(p2.ID, "v1.0.0")); err != nil {
		t.Fatalf("create for p2: %v", err)
	}

	releases, err := releaseRepo.List(ctx, p1.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(releases) != 1 {
		t.Errorf("expected 1 release for p1, got %d", len(releases))
	}
}
