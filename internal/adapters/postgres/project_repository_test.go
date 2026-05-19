//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func baseProject(slug string) domain.Project {
	return domain.Project{
		Slug:           slug,
		Name:           "Test Project",
		GitHubOwner:    "owner",
		GitHubRepo:     "repo",
		DefaultBranch:  "main",
		RootDirectory:  ".",
		DockerfilePath: "Dockerfile",
		BuildContext:   ".",
	}
}

func TestProjectRepository_CreateAndList(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewProjectRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, baseProject("test-project"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID after create")
	}
	if created.Slug != "test-project" {
		t.Errorf("slug: got %q, want %q", created.Slug, "test-project")
	}

	projects, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", projects[0].ID, created.ID)
	}
}

func TestProjectRepository_Create_DuplicateSlug(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewProjectRepository(pool)
	ctx := context.Background()

	if _, err := repo.Create(ctx, baseProject("dup-slug")); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(ctx, baseProject("dup-slug"))
	if err != postgres.ErrProjectSlugExists {
		t.Errorf("expected ErrProjectSlugExists, got %v", err)
	}
}

func TestProjectRepository_List_Empty(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewProjectRepository(pool)
	ctx := context.Background()

	projects, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if projects == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestProjectRepository_Create_FieldsRoundtrip(t *testing.T) {
	pool := setupTestDB(t)
	repo := postgres.NewProjectRepository(pool)
	ctx := context.Background()

	input := domain.Project{
		Slug:           "roundtrip",
		Name:           "Roundtrip Test",
		GitHubOwner:    "my-org",
		GitHubRepo:     "my-repo",
		DefaultBranch:  "develop",
		RootDirectory:  "backend",
		DockerfilePath: "build/Dockerfile",
		BuildContext:   "backend",
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	projects, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1, got %d", len(projects))
	}

	got := projects[0]
	checks := map[string][2]string{
		"Slug":           {got.Slug, created.Slug},
		"Name":           {got.Name, created.Name},
		"GitHubOwner":    {got.GitHubOwner, created.GitHubOwner},
		"GitHubRepo":     {got.GitHubRepo, created.GitHubRepo},
		"DefaultBranch":  {got.DefaultBranch, created.DefaultBranch},
		"RootDirectory":  {got.RootDirectory, created.RootDirectory},
		"DockerfilePath": {got.DockerfilePath, created.DockerfilePath},
		"BuildContext":   {got.BuildContext, created.BuildContext},
	}
	for field, pair := range checks {
		if pair[0] != pair[1] {
			t.Errorf("%s: got %q, want %q", field, pair[0], pair[1])
		}
	}
}
