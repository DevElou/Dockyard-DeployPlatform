package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	"github.com/elouan/dockyard/internal/application/containerlogs"
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

func handleListDeploymentEvents(svc *deploymentapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "deploymentId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		events, err := svc.ListEvents(r.Context(), id)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_events_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": events})
	}
}

func handleGetDeploymentContainerLogs(svc *containerlogs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			httpjson.Error(w, http.StatusServiceUnavailable, "logs_disabled",
				"container logs are not available (DOCKYARD_AGENT_KEY not configured)")
			return
		}

		id, err := requireParam(r, "deploymentId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		tail := 0
		if raw := r.URL.Query().Get("tail"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed <= 0 {
				httpjson.Error(w, http.StatusBadRequest, "invalid_tail",
					"tail must be a positive integer")
				return
			}
			tail = parsed
		}

		resp, err := svc.GetLogs(r.Context(), id, tail)
		if err != nil {
			if errors.Is(err, containerlogs.ErrAgentUnavailable) {
				httpjson.Error(w, http.StatusServiceUnavailable, "agent_unavailable", err.Error())
				return
			}
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, resp)
	}
}
