package httpapi

import (
	"net/http"

	deploymentapp "github.com/elouan/dockyard/internal/application/deployment"
	domainsvc "github.com/elouan/dockyard/internal/application/domainsvc"
	envapp "github.com/elouan/dockyard/internal/application/environment"
	projectapp "github.com/elouan/dockyard/internal/application/project"
	projectserviceapp "github.com/elouan/dockyard/internal/application/projectservice"
	releaseapp "github.com/elouan/dockyard/internal/application/release"
	runtimetargetapp "github.com/elouan/dockyard/internal/application/runtimetarget"
)

type RouterDeps struct {
	ProjectService        *projectapp.Service
	RuntimeTargetService  *runtimetargetapp.Service
	ReleaseService        *releaseapp.Service
	DeploymentService     *deploymentapp.Service
	DomainService         *domainsvc.Service
	ProjectServiceService *projectserviceapp.Service
	EnvironmentService    *envapp.Service
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealth)

	// Projects
	mux.HandleFunc("GET /api/v1/projects", handleListProjects(deps.ProjectService))
	mux.HandleFunc("POST /api/v1/projects", handleCreateProject(deps.ProjectService))
	mux.HandleFunc("GET /api/v1/projects/{id}", handleGetProject(deps.ProjectService))
	mux.HandleFunc("DELETE /api/v1/projects/{id}", handleDeleteProject(deps.ProjectService))
	mux.HandleFunc("GET /api/v1/projects/{id}/runtime-targets", handleListProjectRuntimeTargets(deps.ProjectService))
	mux.HandleFunc("POST /api/v1/projects/{id}/runtime-targets", handleAddProjectRuntimeTarget(deps.ProjectService))

	// Runtime targets
	mux.HandleFunc("GET /api/v1/runtime-targets", handleListRuntimeTargets(deps.RuntimeTargetService))
	mux.HandleFunc("POST /api/v1/runtime-targets", handleCreateRuntimeTarget(deps.RuntimeTargetService))
	mux.HandleFunc("GET /api/v1/runtime-targets/{id}", handleGetRuntimeTarget(deps.RuntimeTargetService))
	mux.HandleFunc("PATCH /api/v1/runtime-targets/{id}/enable", handleEnableRuntimeTarget(deps.RuntimeTargetService))
	mux.HandleFunc("PATCH /api/v1/runtime-targets/{id}/disable", handleDisableRuntimeTarget(deps.RuntimeTargetService))

	// Releases
	mux.HandleFunc("GET /api/v1/projects/{projectId}/releases", handleListReleases(deps.ReleaseService))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/releases", handleCreateRelease(deps.ReleaseService))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/releases/{releaseId}", handleGetRelease(deps.ReleaseService))

	// Deployments
	mux.HandleFunc("GET /api/v1/projects/{projectId}/deployments", handleListDeployments(deps.DeploymentService))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/deployments", handleCreateDeployment(deps.DeploymentService))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/deployments/{deploymentId}", handleGetDeployment(deps.DeploymentService))

	// Domains
	mux.HandleFunc("GET /api/v1/projects/{projectId}/domains", handleListDomains(deps.DomainService))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/domains", handleCreateDomain(deps.DomainService))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/domains/{domainId}", handleGetDomain(deps.DomainService))
	mux.HandleFunc("DELETE /api/v1/projects/{projectId}/domains/{domainId}", handleDeleteDomain(deps.DomainService))

	// Project services
	mux.HandleFunc("GET /api/v1/projects/{projectId}/services", handleListServices(deps.ProjectServiceService))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/services", handleCreateService(deps.ProjectServiceService))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/services/{serviceId}", handleGetService(deps.ProjectServiceService))

	// Environment sets and variables
	mux.HandleFunc("GET /api/v1/projects/{projectId}/environments", handleListEnvironmentSets(deps.EnvironmentService))
	mux.HandleFunc("POST /api/v1/projects/{projectId}/environments", handleCreateEnvironmentSet(deps.EnvironmentService))
	mux.HandleFunc("GET /api/v1/projects/{projectId}/environments/{envId}/variables", handleListVariables(deps.EnvironmentService))
	mux.HandleFunc("PUT /api/v1/projects/{projectId}/environments/{envId}/variables", handleUpsertVariable(deps.EnvironmentService))
	mux.HandleFunc("DELETE /api/v1/projects/{projectId}/environments/{envId}/variables/{varId}", handleDeleteVariable(deps.EnvironmentService))

	return withLogging(mux)
}
