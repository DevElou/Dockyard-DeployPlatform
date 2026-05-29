//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
)

func newOperationEvent(resourceID, phase, msg string) domain.OperationEvent {
	return domain.OperationEvent{
		ResourceType: domain.OperationResourceRelease,
		ResourceID:   resourceID,
		Phase:        phase,
		Level:        domain.OperationLevelInfo,
		Message:      msg,
	}
}

func TestOperationLogRepository_AppendAndList(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p, _ := projectRepo.Create(ctx, baseProject("oplog-list"))

	releaseRepo := postgres.NewReleaseRepository(pool)
	r, err := releaseRepo.Create(ctx, baseRelease(p.ID, "v1.0.0"))
	if err != nil {
		t.Fatalf("create release: %v", err)
	}

	repo := postgres.NewOperationLogRepository(pool)

	ev1, err := repo.Append(ctx, newOperationEvent(r.ID, "queued", "picked up"))
	if err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if ev1.ID == "" || ev1.CreatedAt.IsZero() {
		t.Fatal("expected ID and CreatedAt to be populated")
	}

	withDetails := newOperationEvent(r.ID, "building_image", "build started")
	withDetails.Details = map[string]string{"dockerfile": "Dockerfile"}
	if _, err := repo.Append(ctx, withDetails); err != nil {
		t.Fatalf("append 2: %v", err)
	}

	events, err := repo.List(ctx, domain.OperationResourceRelease, r.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Phase != "queued" || events[1].Phase != "building_image" {
		t.Errorf("events ordered wrong: %+v", events)
	}
	if events[1].Details["dockerfile"] != "Dockerfile" {
		t.Errorf("details lost: %+v", events[1].Details)
	}
}

func TestOperationLogRepository_PruneKeepsLatest(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p, _ := projectRepo.Create(ctx, baseProject("oplog-prune"))

	releaseRepo := postgres.NewReleaseRepository(pool)
	r, _ := releaseRepo.Create(ctx, baseRelease(p.ID, "v1.0.0"))

	// Force tight retention so we can assert the bound is enforced without
	// inserting 1000+ rows in the test.
	repo := postgres.NewOperationLogRepository(pool).WithRetention(3)

	for i := 0; i < 5; i++ {
		ev := newOperationEvent(r.ID, "tick", "evt")
		if _, err := repo.Append(ctx, ev); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	events, err := repo.List(ctx, domain.OperationResourceRelease, r.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events after prune, got %d", len(events))
	}
}

func TestOperationLogRepository_IsolatesResources(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	projectRepo := postgres.NewProjectRepository(pool)
	p, _ := projectRepo.Create(ctx, baseProject("oplog-isolation"))

	releaseRepo := postgres.NewReleaseRepository(pool)
	rA, _ := releaseRepo.Create(ctx, baseRelease(p.ID, "v1.0.0"))
	rB, _ := releaseRepo.Create(ctx, baseRelease(p.ID, "v2.0.0"))

	repo := postgres.NewOperationLogRepository(pool)
	_, _ = repo.Append(ctx, newOperationEvent(rA.ID, "queued", "A"))
	_, _ = repo.Append(ctx, newOperationEvent(rB.ID, "queued", "B"))

	a, _ := repo.List(ctx, domain.OperationResourceRelease, rA.ID)
	b, _ := repo.List(ctx, domain.OperationResourceRelease, rB.ID)
	if len(a) != 1 || len(b) != 1 {
		t.Fatalf("expected one event per release, got A=%d B=%d", len(a), len(b))
	}
	if a[0].Message != "A" || b[0].Message != "B" {
		t.Errorf("cross-talk between resources: A=%v B=%v", a[0], b[0])
	}
}
