package operationlog_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/elouan/dockyard/internal/application/operationlog"
	"github.com/elouan/dockyard/internal/domain"
)

type fakeRepo struct {
	mu        sync.Mutex
	appended  []domain.OperationEvent
	listErr   error
	listResp  []domain.OperationEvent
	appendErr error
}

func (f *fakeRepo) Append(ctx context.Context, ev domain.OperationEvent) (domain.OperationEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.appendErr != nil {
		return domain.OperationEvent{}, f.appendErr
	}
	ev.ID = "fake-id"
	f.appended = append(f.appended, ev)
	return ev, nil
}

func (f *fakeRepo) List(ctx context.Context, resourceType domain.OperationResourceType, resourceID string) ([]domain.OperationEvent, error) {
	return f.listResp, f.listErr
}

func (f *fakeRepo) Prune(ctx context.Context, resourceType domain.OperationResourceType, resourceID string, keep int) error {
	return nil
}

func TestService_Info_AppendsEvent(t *testing.T) {
	repo := &fakeRepo{}
	svc := operationlog.NewService(repo)

	svc.Info(context.Background(), domain.OperationResourceRelease, "rel-1", "building_image", "go", map[string]string{"k": "v"})

	if len(repo.appended) != 1 {
		t.Fatalf("expected 1 appended event, got %d", len(repo.appended))
	}
	ev := repo.appended[0]
	if ev.Level != domain.OperationLevelInfo {
		t.Errorf("expected info level, got %s", ev.Level)
	}
	if ev.Phase != "building_image" || ev.Message != "go" {
		t.Errorf("unexpected event: %+v", ev)
	}
	if ev.Details["k"] != "v" {
		t.Errorf("details lost: %+v", ev.Details)
	}
}

func TestService_LevelMethodsSetCorrectLevel(t *testing.T) {
	repo := &fakeRepo{}
	svc := operationlog.NewService(repo)
	ctx := context.Background()

	svc.Warn(ctx, domain.OperationResourceDeployment, "d", "p", "m", nil)
	svc.Error(ctx, domain.OperationResourceDeployment, "d", "p", "m", nil)
	svc.Success(ctx, domain.OperationResourceDeployment, "d", "p", "m", nil)

	want := []domain.OperationLevel{
		domain.OperationLevelWarn,
		domain.OperationLevelError,
		domain.OperationLevelSuccess,
	}
	if len(repo.appended) != len(want) {
		t.Fatalf("expected %d events, got %d", len(want), len(repo.appended))
	}
	for i, w := range want {
		if repo.appended[i].Level != w {
			t.Errorf("event %d: expected level %s, got %s", i, w, repo.appended[i].Level)
		}
	}
}

func TestService_Record_SwallowsRepoErrors(t *testing.T) {
	// Observability must not break the workflow it observes — Record returns
	// the error to callers that want it, but the Info/Warn/... helpers swallow.
	repo := &fakeRepo{appendErr: errors.New("db down")}
	svc := operationlog.NewService(repo)

	// helpers should not panic and should not propagate the failure
	svc.Info(context.Background(), domain.OperationResourceRelease, "rel", "queued", "m", nil)

	// Record itself does propagate
	err := svc.Record(context.Background(), domain.OperationEvent{
		ResourceType: domain.OperationResourceRelease,
		ResourceID:   "rel",
		Phase:        "queued",
		Level:        domain.OperationLevelInfo,
		Message:      "m",
	})
	if err == nil {
		t.Fatal("expected error from Record when repo fails")
	}
}

func TestService_ListForResource(t *testing.T) {
	repo := &fakeRepo{
		listResp: []domain.OperationEvent{{ID: "1"}, {ID: "2"}},
	}
	svc := operationlog.NewService(repo)

	got, err := svc.ListForResource(context.Background(), domain.OperationResourceRelease, "rel-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
}
