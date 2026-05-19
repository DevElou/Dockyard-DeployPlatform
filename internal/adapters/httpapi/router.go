package httpapi

import (
	"net/http"

	deploymentapp "github.com/elouan/dockyard/internal/application/deployment"
	domainsvc "github.com/elouan/dockyard/internal/application/domainsvc"
	projectapp "github.com/elouan/dockyard/internal/application/project"
	releaseapp "github.com/elouan/dockyard/internal/application/release"
	runtimetargetapp "github.com/elouan/dockyard/internal/application/runtimetarget"
)

type RouterDeps struct {
	ProjectService       *projectapp.Service
	RuntimeTargetService *runtimetargetapp.Service
	ReleaseService       *releaseapp.Service
	DeploymentService    *deploymentapp.Service
	DomainService        *domainsvc.Service
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

	return withLogging(mux)
}
