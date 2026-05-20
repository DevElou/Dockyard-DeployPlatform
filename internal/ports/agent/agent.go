package agent

import (
	"context"

	"github.com/elouan/dockyard/internal/ports/runtime"
)

type DeployRequest struct {
	Spec runtime.DeploymentSpec
}

type DeployResponse struct {
	Accepted     bool
	DeploymentID string
}

type StatusResponse struct {
	DeploymentID string
	Result       runtime.DeploymentResult
}

type Client interface {
	Deploy(ctx context.Context, request DeployRequest) (DeployResponse, error)
	GetStatus(ctx context.Context, deploymentID string) (StatusResponse, error)
	Remove(ctx context.Context, deploymentID string) error
}
