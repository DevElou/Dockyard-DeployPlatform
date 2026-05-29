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

type LogsRequest struct {
	DeploymentID string
	Tail         int
}

type LogsResponse struct {
	DeploymentID string `json:"deploymentId"`
	ContainerID  string `json:"containerId"`
	Tail         int    `json:"tail"`
	Logs         string `json:"logs"`
}

type Client interface {
	Deploy(ctx context.Context, request DeployRequest) (DeployResponse, error)
	GetStatus(ctx context.Context, deploymentID string) (StatusResponse, error)
	GetLogs(ctx context.Context, request LogsRequest) (LogsResponse, error)
	Remove(ctx context.Context, deploymentID string) error
}
