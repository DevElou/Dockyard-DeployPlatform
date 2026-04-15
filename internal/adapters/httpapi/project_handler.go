package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	projectapp "github.com/elouan/dockyard/internal/application/project"
)

func handleListProjects(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := service.List(r.Context())
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_projects_failed", err.Error())
			return
		}

		httpjson.Write(w, http.StatusOK, map[string]any{
			"items": projects,
		})
	}
}

func handleCreateProject(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input projectapp.CreateProjectInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		project, err := service.Create(r.Context(), input)
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "create_project_failed", err.Error())
			return
		}

		httpjson.Write(w, http.StatusCreated, project)
	}
}
