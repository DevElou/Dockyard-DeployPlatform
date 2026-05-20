package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/runtime"
)

type Worker struct {
	deployments    repository.DeploymentRepository
	releases       repository.ReleaseRepository
	projects       repository.ProjectRepository
	runtimeTargets repository.RuntimeTargetRepository
	agentClients   AgentClientFactory
	inFlight       sync.Map // deploymentID → struct{}
}

// AgentClientFactory returns an agent.Client given a RuntimeTarget.
type AgentClientFactory func(target domain.RuntimeTarget) (agent.Client, error)

func NewWorker(
	deployments repository.DeploymentRepository,
	releases repository.ReleaseRepository,
	projects repository.ProjectRepository,
	runtimeTargets repository.RuntimeTargetRepository,
	factory AgentClientFactory,
) *Worker {
	return &Worker{
		deployments:    deployments,
		releases:       releases,
		projects:       projects,
		runtimeTargets: runtimeTargets,
		agentClients:   factory,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("orchestrator-worker: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("orchestrator-worker: stopping")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	pending, err := w.deployments.ListByStatus(ctx, domain.DeploymentStatusPending)
	if err != nil {
		log.Printf("orchestrator-worker: list pending deployments: %v", err)
		return
	}

	for _, d := range pending {
		if _, loaded := w.inFlight.LoadOrStore(d.ID, struct{}{}); loaded {
			continue // already processing
		}
		go func(dep domain.Deployment) {
			defer w.inFlight.Delete(dep.ID)
			w.processDeployment(ctx, dep)
		}(d)
	}
}

func (w *Worker) processDeployment(ctx context.Context, d domain.Deployment) {
	log.Printf("orchestrator-worker: processing deployment %s", d.ID)

	now := time.Now()
	if err := w.deployments.UpdateStatus(ctx, d.ID, domain.DeploymentStatusDeploying, &now, nil); err != nil {
		log.Printf("orchestrator-worker: claim deployment %s: %v", d.ID, err)
		return
	}

	spec, err := w.buildSpec(ctx, d)
	if err != nil {
		log.Printf("orchestrator-worker: build spec for deployment %s: %v", d.ID, err)
		w.fail(d.ID)
		return
	}

	target, err := w.runtimeTargets.GetByID(ctx, d.RuntimeTargetID)
	if err != nil {
		log.Printf("orchestrator-worker: get runtime target %s: %v", d.RuntimeTargetID, err)
		w.fail(d.ID)
		return
	}

	agentClient, err := w.agentClients(target)
	if err != nil {
		log.Printf("orchestrator-worker: get agent client for target %s: %v", target.ID, err)
		w.fail(d.ID)
		return
	}

	resp, err := agentClient.Deploy(ctx, agent.DeployRequest{Spec: spec})
	if err != nil {
		log.Printf("orchestrator-worker: send deploy request for %s: %v", d.ID, err)
		w.fail(d.ID)
		return
	}

	if !resp.Accepted {
		log.Printf("orchestrator-worker: deploy %s not accepted by agent", d.ID)
		w.fail(d.ID)
		return
	}

	w.waitForCompletion(ctx, d.ID, agentClient)
}

func (w *Worker) waitForCompletion(ctx context.Context, deploymentID string, agentClient agent.Client) {
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			log.Printf("orchestrator-worker: timeout waiting for deployment %s: %v", deploymentID, pollCtx.Err())
			w.fail(deploymentID)
			return
		case <-ticker.C:
			statusResp, err := agentClient.GetStatus(pollCtx, deploymentID)
			if err != nil {
				log.Printf("orchestrator-worker: get status for deployment %s: %v", deploymentID, err)
				continue
			}

			switch statusResp.Result.Status {
			case domain.DeploymentStatusHealthy:
				now := time.Now()
				if err := w.deployments.UpdateStatus(ctx, deploymentID, domain.DeploymentStatusHealthy, nil, &now); err != nil {
					log.Printf("orchestrator-worker: persist healthy status for %s: %v", deploymentID, err)
				}
				log.Printf("orchestrator-worker: deployment %s is healthy", deploymentID)
				return
			case domain.DeploymentStatusFailed:
				log.Printf("orchestrator-worker: deployment %s failed: %s", deploymentID, statusResp.Result.Message)
				w.fail(deploymentID)
				return
			default:
				// still deploying, keep polling
			}
		}
	}
}

func (w *Worker) buildSpec(ctx context.Context, d domain.Deployment) (runtime.DeploymentSpec, error) {
	release, err := w.releases.GetByID(ctx, d.ReleaseID)
	if err != nil {
		return runtime.DeploymentSpec{}, err
	}

	project, err := w.projects.GetByID(ctx, d.ProjectID)
	if err != nil {
		return runtime.DeploymentSpec{}, err
	}

	// TODO: populate Service and Environment fields once project_services and
	// environment_variables repositories are exposed. Currently containers start
	// without port bindings or env vars.
	return runtime.DeploymentSpec{
		DeploymentID: d.ID,
		ProjectID:    d.ProjectID,
		ProjectSlug:  project.Slug,
		ReleaseID:    d.ReleaseID,
		Image: runtime.ImageRef{
			Repository: release.ImageRepository,
			Tag:        release.ImageTag,
			Digest:     release.ImageDigest,
		},
		Strategy: d.Strategy,
	}, nil
}

// fail writes a failed terminal status using a fresh context so it succeeds
// even if the caller's context is already cancelled (e.g. on shutdown or timeout).
func (w *Worker) fail(deploymentID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	now := time.Now()
	if err := w.deployments.UpdateStatus(ctx, deploymentID, domain.DeploymentStatusFailed, nil, &now); err != nil {
		log.Printf("orchestrator-worker: fail deployment %s: %v", deploymentID, err)
	}
}
