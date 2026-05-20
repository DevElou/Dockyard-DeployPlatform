package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	"github.com/elouan/dockyard/internal/adapters/postgres"
	projectapp "github.com/elouan/dockyard/internal/application/project"
)

func handleListProjects(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := service.List(r.Context())
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_projects_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": projects})
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
			if err == postgres.ErrProjectSlugExists {
				httpjson.Error(w, http.StatusConflict, "conflict", err.Error())
				return
			}
			httpjson.Error(w, http.StatusBadRequest, "create_project_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusCreated, project)
	}
}

func handleGetProject(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		project, err := service.GetByID(r.Context(), id)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, project)
	}
}

func handleDeleteProject(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		if err := service.Archive(r.Context(), id); err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}

func handleListProjectRuntimeTargets(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		targets, err := service.ListRuntimeTargets(r.Context(), id)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": targets})
	}
}

func handleAddProjectRuntimeTarget(service *projectapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input projectapp.AddRuntimeTargetInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		if err := service.AddRuntimeTarget(r.Context(), id, input); err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}
