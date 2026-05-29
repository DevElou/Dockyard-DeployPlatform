// Package containerlogs fetches recent runtime logs from the deploy-agent on
// demand. Logs are not persisted in Dockyard's database — only fetched live
// when the user opens the deployment's "Container logs" tab.
package containerlogs

import (
	"context"
	"errors"
	"fmt"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
	"github.com/elouan/dockyard/internal/ports/repository"
)

const (
	defaultTail = 300
	maxTail     = 5000
)

// AgentClientFactory builds an agent.Client for the given runtime target.
type AgentClientFactory func(target domain.RuntimeTarget) (agent.Client, error)

type Service struct {
	deployments    repository.DeploymentRepository
	runtimeTargets repository.RuntimeTargetRepository
	agentClients   AgentClientFactory
}

func NewService(
	deployments repository.DeploymentRepository,
	runtimeTargets repository.RuntimeTargetRepository,
	factory AgentClientFactory,
) *Service {
	return &Service{
		deployments:    deployments,
		runtimeTargets: runtimeTargets,
		agentClients:   factory,
	}
}

// ErrAgentUnavailable is returned when the factory cannot produce a client
// (e.g. the runtime target has no endpoint, or the API was started without
// the DOCKYARD_AGENT_KEY).
var ErrAgentUnavailable = errors.New("deploy-agent not reachable")

func (s *Service) GetLogs(ctx context.Context, deploymentID string, tail int) (agent.LogsResponse, error) {
	if tail <= 0 {
		tail = defaultTail
	}
	if tail > maxTail {
		tail = maxTail
	}

	dep, err := s.deployments.GetByID(ctx, deploymentID)
	if err != nil {
		return agent.LogsResponse{}, err
	}

	target, err := s.runtimeTargets.GetByID(ctx, dep.RuntimeTargetID)
	if err != nil {
		return agent.LogsResponse{}, fmt.Errorf("get runtime target: %w", err)
	}

	if s.agentClients == nil {
		return agent.LogsResponse{}, ErrAgentUnavailable
	}
	client, err := s.agentClients(target)
	if err != nil {
		return agent.LogsResponse{}, fmt.Errorf("%w: %v", ErrAgentUnavailable, err)
	}

	return client.GetLogs(ctx, agent.LogsRequest{DeploymentID: deploymentID, Tail: tail})
}
