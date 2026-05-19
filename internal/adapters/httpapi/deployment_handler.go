package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	deploymentapp "github.com/elouan/dockyard/internal/application/deployment"
)

func handleListDeployments(svc *deploymentapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		deployments, err := svc.List(r.Context(), projectID)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_deployments_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": deployments})
	}
}

func handleCreateDeployment(svc *deploymentapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input deploymentapp.CreateDeploymentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		deployment, err := svc.Create(r.Context(), projectID, input)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, deployment)
	}
}

func handleGetDeployment(svc *deploymentapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "deploymentId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		deployment, err := svc.GetByID(r.Context(), id)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, deployment)
	}
}
