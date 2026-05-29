package containerlogs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elouan/dockyard/internal/application/containerlogs"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
)

type fakeDeploymentRepo struct {
	dep domain.Deployment
	err error
}

func (f *fakeDeploymentRepo) List(ctx context.Context, projectID string) ([]domain.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentRepo) ListByStatus(ctx context.Context, status domain.DeploymentStatus) ([]domain.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentRepo) Create(ctx context.Context, d domain.Deployment) (domain.Deployment, error) {
	return d, nil
}
func (f *fakeDeploymentRepo) GetByID(ctx context.Context, id string) (domain.Deployment, error) {
	return f.dep, f.err
}
func (f *fakeDeploymentRepo) UpdateStatus(ctx context.Context, id string, status domain.DeploymentStatus, startedAt, finishedAt *time.Time) error {
	return nil
}

type fakeRuntimeRepo struct {
	target domain.RuntimeTarget
	err    error
}

func (f *fakeRuntimeRepo) List(ctx context.Context) ([]domain.RuntimeTarget, error) {
	return nil, nil
}
func (f *fakeRuntimeRepo) Create(ctx context.Context, t domain.RuntimeTarget) (domain.RuntimeTarget, error) {
	return t, nil
}
func (f *fakeRuntimeRepo) GetByID(ctx context.Context, id string) (domain.RuntimeTarget, error) {
	return f.target, f.err
}
func (f *fakeRuntimeRepo) SetEnabled(ctx context.Context, id string, enabled bool) error {
	return nil
}

type fakeAgentClient struct {
	gotReq  agent.LogsRequest
	respLog agent.LogsResponse
	err     error
}

func (f *fakeAgentClient) Deploy(ctx context.Context, req agent.DeployRequest) (agent.DeployResponse, error) {
	return agent.DeployResponse{}, nil
}
func (f *fakeAgentClient) GetStatus(ctx context.Context, id string) (agent.StatusResponse, error) {
	return agent.StatusResponse{}, nil
}
func (f *fakeAgentClient) GetLogs(ctx context.Context, req agent.LogsRequest) (agent.LogsResponse, error) {
	f.gotReq = req
	return f.respLog, f.err
}
func (f *fakeAgentClient) Remove(ctx context.Context, id string) error { return nil }

func TestService_GetLogs_ProxiesToAgent(t *testing.T) {
	deps := &fakeDeploymentRepo{dep: domain.Deployment{ID: "dep-1", RuntimeTargetID: "rt-1"}}
	rts := &fakeRuntimeRepo{target: domain.RuntimeTarget{ID: "rt-1", Endpoint: "http://agent.local"}}
	client := &fakeAgentClient{respLog: agent.LogsResponse{
		DeploymentID: "dep-1",
		ContainerID:  "c1",
		Tail:         123,
		Logs:         "hello",
	}}
	factory := func(domain.RuntimeTarget) (agent.Client, error) { return client, nil }

	svc := containerlogs.NewService(deps, rts, factory)

	got, err := svc.GetLogs(context.Background(), "dep-1", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ContainerID != "c1" || got.Logs != "hello" {
		t.Errorf("unexpected response: %+v", got)
	}
	if client.gotReq.DeploymentID != "dep-1" || client.gotReq.Tail != 123 {
		t.Errorf("agent client got wrong request: %+v", client.gotReq)
	}
}

func TestService_GetLogs_DefaultsTail(t *testing.T) {
	deps := &fakeDeploymentRepo{dep: domain.Deployment{ID: "d", RuntimeTargetID: "rt"}}
	rts := &fakeRuntimeRepo{target: domain.RuntimeTarget{ID: "rt", Endpoint: "x"}}
	client := &fakeAgentClient{}
	factory := func(domain.RuntimeTarget) (agent.Client, error) { return client, nil }

	svc := containerlogs.NewService(deps, rts, factory)
	if _, err := svc.GetLogs(context.Background(), "d", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.gotReq.Tail != 300 {
		t.Errorf("expected default tail 300, got %d", client.gotReq.Tail)
	}
}

func TestService_GetLogs_CapsTail(t *testing.T) {
	deps := &fakeDeploymentRepo{dep: domain.Deployment{ID: "d", RuntimeTargetID: "rt"}}
	rts := &fakeRuntimeRepo{target: domain.RuntimeTarget{ID: "rt", Endpoint: "x"}}
	client := &fakeAgentClient{}
	factory := func(domain.RuntimeTarget) (agent.Client, error) { return client, nil }

	svc := containerlogs.NewService(deps, rts, factory)
	if _, err := svc.GetLogs(context.Background(), "d", 999999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.gotReq.Tail != 5000 {
		t.Errorf("expected capped tail 5000, got %d", client.gotReq.Tail)
	}
}

func TestService_GetLogs_FactoryError(t *testing.T) {
	deps := &fakeDeploymentRepo{dep: domain.Deployment{ID: "d", RuntimeTargetID: "rt"}}
	rts := &fakeRuntimeRepo{target: domain.RuntimeTarget{ID: "rt", Endpoint: "x"}}
	factory := func(domain.RuntimeTarget) (agent.Client, error) { return nil, errors.New("no endpoint") }

	svc := containerlogs.NewService(deps, rts, factory)
	_, err := svc.GetLogs(context.Background(), "d", 50)
	if !errors.Is(err, containerlogs.ErrAgentUnavailable) {
		t.Fatalf("expected ErrAgentUnavailable, got %v", err)
	}
}
