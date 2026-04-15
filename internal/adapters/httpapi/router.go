package httpapi

import (
	"net/http"

	projectapp "github.com/elouan/dockyard/internal/application/project"
)

type RouterDeps struct {
	ProjectService *projectapp.Service
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /api/v1/projects", handleListProjects(deps.ProjectService))
	mux.HandleFunc("POST /api/v1/projects", handleCreateProject(deps.ProjectService))

	return withLogging(mux)
}
