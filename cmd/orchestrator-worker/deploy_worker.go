package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/runtime"
)

// AgentClientFactory returns an agent.Client given a RuntimeTarget.
type AgentClientFactory func(target domain.RuntimeTarget) (agent.Client, error)

type DeployWorker struct {
	deployments     repository.DeploymentRepository
	releases        repository.ReleaseRepository
	projects        repository.ProjectRepository
	runtimeTargets  repository.RuntimeTargetRepository
	projectServices repository.ProjectServiceRepository
	envSets         repository.EnvironmentSetRepository
	envVars         repository.EnvironmentVariableRepository
	agentClients    AgentClientFactory
	inFlight        sync.Map // deploymentID → struct{}
}

func NewDeployWorker(
	deployments repository.DeploymentRepository,
	releases repository.ReleaseRepository,
	projects repository.ProjectRepository,
	runtimeTargets repository.RuntimeTargetRepository,
	projectServices repository.ProjectServiceRepository,
	envSets repository.EnvironmentSetRepository,
	envVars repository.EnvironmentVariableRepository,
	factory AgentClientFactory,
) *DeployWorker {
	return &DeployWorker{
		deployments:     deployments,
		releases:        releases,
		projects:        projects,
		runtimeTargets:  runtimeTargets,
		projectServices: projectServices,
		envSets:         envSets,
		envVars:         envVars,
		agentClients:    factory,
	}
}

func (w *DeployWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("deploy-worker: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("deploy-worker: stopping")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *DeployWorker) tick(ctx context.Context) {
	pending, err := w.deployments.ListByStatus(ctx, domain.DeploymentStatusPending)
	if err != nil {
		log.Printf("deploy-worker: list pending deployments: %v", err)
		return
	}

	for _, d := range pending {
		if _, loaded := w.inFlight.LoadOrStore(d.ID, struct{}{}); loaded {
			continue
		}
		go func(dep domain.Deployment) {
			defer w.inFlight.Delete(dep.ID)
			w.processDeployment(ctx, dep)
		}(d)
	}
}

func (w *DeployWorker) processDeployment(ctx context.Context, d domain.Deployment) {
	log.Printf("deploy-worker: processing deployment %s", d.ID)

	now := time.Now()
	if err := w.deployments.UpdateStatus(ctx, d.ID, domain.DeploymentStatusDeploying, &now, nil); err != nil {
		log.Printf("deploy-worker: claim deployment %s: %v", d.ID, err)
		return
	}

	spec, err := w.buildSpec(ctx, d)
	if err != nil {
		log.Printf("deploy-worker: build spec for deployment %s: %v", d.ID, err)
		w.fail(d.ID)
		return
	}

	target, err := w.runtimeTargets.GetByID(ctx, d.RuntimeTargetID)
	if err != nil {
		log.Printf("deploy-worker: get runtime target %s: %v", d.RuntimeTargetID, err)
		w.fail(d.ID)
		return
	}

	agentClient, err := w.agentClients(target)
	if err != nil {
		log.Printf("deploy-worker: get agent client for target %s: %v", target.ID, err)
		w.fail(d.ID)
		return
	}

	resp, err := agentClient.Deploy(ctx, agent.DeployRequest{Spec: spec})
	if err != nil {
		log.Printf("deploy-worker: send deploy request for %s: %v", d.ID, err)
		w.fail(d.ID)
		return
	}

	if !resp.Accepted {
		log.Printf("deploy-worker: deploy %s not accepted by agent", d.ID)
		w.fail(d.ID)
		return
	}

	w.waitForCompletion(ctx, d.ID, agentClient)
}

func (w *DeployWorker) waitForCompletion(ctx context.Context, deploymentID string, agentClient agent.Client) {
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			log.Printf("deploy-worker: timeout waiting for deployment %s: %v", deploymentID, pollCtx.Err())
			w.fail(deploymentID)
			return
		case <-ticker.C:
			statusResp, err := agentClient.GetStatus(pollCtx, deploymentID)
			if err != nil {
				log.Printf("deploy-worker: get status for deployment %s: %v", deploymentID, err)
				continue
			}

			switch statusResp.Result.Status {
			case domain.DeploymentStatusHealthy:
				now := time.Now()
				if err := w.deployments.UpdateStatus(ctx, deploymentID, domain.DeploymentStatusHealthy, nil, &now); err != nil {
					log.Printf("deploy-worker: persist healthy status for %s: %v", deploymentID, err)
				}
				log.Printf("deploy-worker: deployment %s is healthy", deploymentID)
				return
			case domain.DeploymentStatusFailed:
				log.Printf("deploy-worker: deployment %s failed: %s", deploymentID, statusResp.Result.Message)
				w.fail(deploymentID)
				return
			}
		}
	}
}

func (w *DeployWorker) buildSpec(ctx context.Context, d domain.Deployment) (runtime.DeploymentSpec, error) {
	release, err := w.releases.GetByID(ctx, d.ReleaseID)
	if err != nil {
		return runtime.DeploymentSpec{}, fmt.Errorf("get release: %w", err)
	}

	if release.BuildStatus != domain.BuildStatusSucceeded {
		return runtime.DeploymentSpec{}, fmt.Errorf("release %s build not ready (status: %s)", release.ID, release.BuildStatus)
	}

	project, err := w.projects.GetByID(ctx, d.ProjectID)
	if err != nil {
		return runtime.DeploymentSpec{}, fmt.Errorf("get project: %w", err)
	}

	svcSpec, err := w.resolveServiceSpec(ctx, d)
	if err != nil {
		return runtime.DeploymentSpec{}, fmt.Errorf("resolve service spec: %w", err)
	}

	envVars, err := w.resolveEnvVars(ctx, d)
	if err != nil {
		return runtime.DeploymentSpec{}, fmt.Errorf("resolve env vars: %w", err)
	}

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
		Service:     svcSpec,
		Environment: envVars,
		Strategy:    d.Strategy,
	}, nil
}

func (w *DeployWorker) resolveServiceSpec(ctx context.Context, d domain.Deployment) (runtime.ServiceSpec, error) {
	if d.ProjectServiceID != nil {
		svc, err := w.projectServices.GetByID(ctx, *d.ProjectServiceID)
		if err != nil {
			return runtime.ServiceSpec{}, err
		}
		return runtime.ServiceSpec{
			Name:            svc.Name,
			InternalPort:    svc.ContainerPort,
			HealthcheckPath: svc.HealthcheckPath,
			HealthcheckPort: svc.HealthcheckPort,
		}, nil
	}

	svc, err := w.projectServices.GetDefaultForProject(ctx, d.ProjectID)
	if err != nil {
		if isNotFound(err) {
			return runtime.ServiceSpec{}, nil
		}
		return runtime.ServiceSpec{}, err
	}
	return runtime.ServiceSpec{
		Name:            svc.Name,
		InternalPort:    svc.ContainerPort,
		HealthcheckPath: svc.HealthcheckPath,
		HealthcheckPort: svc.HealthcheckPort,
	}, nil
}

func (w *DeployWorker) resolveEnvVars(ctx context.Context, d domain.Deployment) ([]domain.EnvironmentVariable, error) {
	var setID string
	if d.EnvironmentSetID != nil {
		setID = *d.EnvironmentSetID
	} else {
		set, err := w.envSets.GetDefaultForProject(ctx, d.ProjectID)
		if err != nil {
			if isNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		setID = set.ID
	}

	return w.envVars.ListBySet(ctx, setID)
}

func isNotFound(err error) bool {
	return errors.Is(err, postgres.ErrProjectServiceNotFound) || errors.Is(err, postgres.ErrEnvironmentSetNotFound)
}

func (w *DeployWorker) fail(deploymentID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	now := time.Now()
	if err := w.deployments.UpdateStatus(ctx, deploymentID, domain.DeploymentStatusFailed, nil, &now); err != nil {
		log.Printf("deploy-worker: fail deployment %s: %v", deploymentID, err)
	}
}
