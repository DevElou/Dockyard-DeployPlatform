package agent

import "github.com/elouan/dockyard/internal/ports/runtime"

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
	Deploy(request DeployRequest) (DeployResponse, error)
	GetStatus(deploymentID string) (StatusResponse, error)
	Remove(deploymentID string) error
}
