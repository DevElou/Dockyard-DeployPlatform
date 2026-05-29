package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/application/operationlog"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/routing"
	"github.com/elouan/dockyard/internal/ports/runtime"
)

// Deployment phases — kept narrow so the UI timeline reads naturally.
const (
	phaseDeployQueued          = "queued"
	phaseDeployBuildingSpec    = "building_spec"
	phaseDeployContactingAgent = "contacting_agent"
	phaseDeployHealthCheck     = "health_check"
	phaseDeployRouting         = "routing"
	phaseDeployHealthy         = "healthy"
	phaseDeployFailed          = "failed"
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
	domains         repository.DomainRepository
	routing         routing.Provider
	agentClients    AgentClientFactory
	events          *operationlog.Service
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
	domains repository.DomainRepository,
	routingProvider routing.Provider,
	factory AgentClientFactory,
	events *operationlog.Service,
) *DeployWorker {
	return &DeployWorker{
		deployments:     deployments,
		releases:        releases,
		projects:        projects,
		runtimeTargets:  runtimeTargets,
		projectServices: projectServices,
		envSets:         envSets,
		envVars:         envVars,
		domains:         domains,
		routing:         routingProvider,
		agentClients:    factory,
		events:          events,
	}
}

func (w *DeployWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var wg sync.WaitGroup
	log.Println("deploy-worker: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("deploy-worker: stopping, draining in-flight deployments")
			wg.Wait()
			return
		case <-ticker.C:
			w.tick(ctx, &wg)
		}
	}
}

func (w *DeployWorker) tick(ctx context.Context, wg *sync.WaitGroup) {
	pending, err := w.deployments.ListByStatus(ctx, domain.DeploymentStatusPending)
	if err != nil {
		log.Printf("deploy-worker: list pending deployments: %v", err)
		return
	}

	for _, d := range pending {
		if _, loaded := w.inFlight.LoadOrStore(d.ID, struct{}{}); loaded {
			continue
		}
		wg.Add(1)
		go func(dep domain.Deployment) {
			defer wg.Done()
			defer w.inFlight.Delete(dep.ID)
			w.processDeployment(ctx, dep)
		}(d)
	}
}

func (w *DeployWorker) processDeployment(ctx context.Context, d domain.Deployment) {
	log.Printf("deploy-worker: processing deployment %s", d.ID)
	w.events.Info(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployQueued,
		"deploy worker picked up deployment", nil)

	now := time.Now()
	if err := w.deployments.UpdateStatus(ctx, d.ID, domain.DeploymentStatusDeploying, &now, nil); err != nil {
		log.Printf("deploy-worker: claim deployment %s: %v", d.ID, err)
		w.events.Error(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployFailed,
			"failed to mark deployment as deploying", map[string]string{"error": err.Error()})
		return
	}

	w.events.Info(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployBuildingSpec,
		"resolving release, service spec and environment", nil)

	spec, err := w.buildSpec(ctx, d)
	if err != nil {
		log.Printf("deploy-worker: build spec for deployment %s: %v", d.ID, err)
		w.events.Error(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployFailed,
			"failed to build deployment spec", map[string]string{"error": truncate(err.Error(), 4000)})
		w.fail(d.ID)
		return
	}

	target, err := w.runtimeTargets.GetByID(ctx, d.RuntimeTargetID)
	if err != nil {
		log.Printf("deploy-worker: get runtime target %s: %v", d.RuntimeTargetID, err)
		w.events.Error(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployFailed,
			"runtime target not found", map[string]string{"error": err.Error()})
		w.fail(d.ID)
		return
	}

	agentClient, err := w.agentClients(target)
	if err != nil {
		log.Printf("deploy-worker: get agent client for target %s: %v", target.ID, err)
		w.events.Error(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployFailed,
			"failed to build agent client", map[string]string{"error": err.Error()})
		w.fail(d.ID)
		return
	}

	w.events.Info(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployContactingAgent,
		"sending deployment spec to agent",
		map[string]string{"target": target.Slug, "endpoint": target.Endpoint, "image": specImageRef(spec)})

	resp, err := agentClient.Deploy(ctx, agent.DeployRequest{Spec: spec})
	if err != nil {
		log.Printf("deploy-worker: send deploy request for %s: %v", d.ID, err)
		w.events.Error(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployFailed,
			"agent unreachable", map[string]string{"error": err.Error()})
		w.fail(d.ID)
		return
	}

	if !resp.Accepted {
		log.Printf("deploy-worker: deploy %s not accepted by agent", d.ID)
		w.events.Error(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployFailed,
			"deployment not accepted by agent", nil)
		w.fail(d.ID)
		return
	}

	w.events.Info(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployHealthCheck,
		"polling agent for container health", nil)

	if w.waitForCompletion(ctx, d.ID, agentClient) {
		w.configureRouting(ctx, d, target)
	}
}

func (w *DeployWorker) waitForCompletion(ctx context.Context, deploymentID string, agentClient agent.Client) bool {
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			log.Printf("deploy-worker: timeout waiting for deployment %s: %v", deploymentID, pollCtx.Err())
			w.events.Error(ctx, domain.OperationResourceDeployment, deploymentID, phaseDeployFailed,
				"timeout waiting for healthy container", map[string]string{"error": pollCtx.Err().Error()})
			w.fail(deploymentID)
			return false
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
				w.events.Success(ctx, domain.OperationResourceDeployment, deploymentID, phaseDeployHealthy,
					"container is healthy",
					map[string]string{"containerId": statusResp.Result.ContainerID})
				return true
			case domain.DeploymentStatusFailed:
				log.Printf("deploy-worker: deployment %s failed: %s", deploymentID, statusResp.Result.Message)
				w.events.Error(ctx, domain.OperationResourceDeployment, deploymentID, phaseDeployFailed,
					"deployment reported failure",
					map[string]string{"agentMessage": truncate(statusResp.Result.Message, 4000)})
				w.fail(deploymentID)
				return false
			}
		}
	}
}

func (w *DeployWorker) configureRouting(ctx context.Context, d domain.Deployment, target domain.RuntimeTarget) {
	svc, err := w.resolveService(ctx, d)
	if err != nil {
		log.Printf("deploy-worker: routing: resolve service for deployment %s: %v", d.ID, err)
		w.events.Warn(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployRouting,
			"failed to resolve service for routing", map[string]string{"error": err.Error()})
		return
	}
	if svc == nil || !svc.RoutingEnabled || svc.ContainerPort == 0 {
		return
	}

	u, err := url.Parse(target.Endpoint)
	if err != nil || u.Hostname() == "" {
		log.Printf("deploy-worker: routing: parse endpoint %q: %v", target.Endpoint, err)
		w.events.Warn(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployRouting,
			"could not parse target endpoint", map[string]string{"endpoint": target.Endpoint})
		return
	}
	forwardHost := u.Hostname()

	domains, err := w.domains.ListByProjectService(ctx, svc.ID)
	if err != nil {
		log.Printf("deploy-worker: routing: list domains for service %s: %v", svc.ID, err)
		w.events.Warn(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployRouting,
			"failed to list domains for service", map[string]string{"error": err.Error()})
		return
	}

	for _, dom := range domains {
		req := routing.RouteRequest{
			Hostname:    dom.Hostname,
			ForwardHost: forwardHost,
			TargetPort:  svc.ContainerPort,
			TLS:         dom.TLSEnabled,
		}
		if err := w.routing.EnsureRoute(req); err != nil {
			log.Printf("deploy-worker: routing: ensure route %s: %v", dom.Hostname, err)
			_ = w.domains.UpdateStatus(ctx, dom.ID, domain.DomainStatusFailed)
			w.events.Warn(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployRouting,
				"failed to configure route",
				map[string]string{"hostname": dom.Hostname, "error": err.Error()})
		} else {
			_ = w.domains.UpdateStatus(ctx, dom.ID, domain.DomainStatusReady)
			w.events.Info(ctx, domain.OperationResourceDeployment, d.ID, phaseDeployRouting,
				"route configured",
				map[string]string{"hostname": dom.Hostname, "forwardHost": forwardHost})
		}
	}
}

func (w *DeployWorker) resolveService(ctx context.Context, d domain.Deployment) (*domain.ProjectService, error) {
	if d.ProjectServiceID != nil {
		svc, err := w.projectServices.GetByID(ctx, *d.ProjectServiceID)
		if err != nil {
			return nil, err
		}
		return &svc, nil
	}
	svc, err := w.projectServices.GetDefaultForProject(ctx, d.ProjectID)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &svc, nil
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
			return runtime.ServiceSpec{}, fmt.Errorf("get service by id: %w", err)
		}
		return toServiceSpec(svc), nil
	}

	svc, err := w.projectServices.GetDefaultForProject(ctx, d.ProjectID)
	if err != nil {
		if isNotFound(err) {
			return runtime.ServiceSpec{}, nil
		}
		return runtime.ServiceSpec{}, fmt.Errorf("get default service: %w", err)
	}
	return toServiceSpec(svc), nil
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
			return nil, fmt.Errorf("get default environment set: %w", err)
		}
		setID = set.ID
	}

	vars, err := w.envVars.ListBySet(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("list env vars for set %s: %w", setID, err)
	}
	return vars, nil
}

func toServiceSpec(svc domain.ProjectService) runtime.ServiceSpec {
	return runtime.ServiceSpec{
		Name:            svc.Name,
		InternalPort:    svc.ContainerPort,
		HealthcheckPath: svc.HealthcheckPath,
		HealthcheckPort: svc.HealthcheckPort,
	}
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

func specImageRef(spec runtime.DeploymentSpec) string {
	if spec.Image.Digest != "" {
		return spec.Image.Repository + "@" + spec.Image.Digest
	}
	if spec.Image.Tag != "" {
		return spec.Image.Repository + ":" + spec.Image.Tag
	}
	return spec.Image.Repository
}
