package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	projectserviceapp "github.com/elouan/dockyard/internal/application/projectservice"
)

func handleListServices(svc *projectserviceapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		services, err := svc.List(r.Context(), projectID)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_services_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": services})
	}
}

func handleCreateService(svc *projectserviceapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input projectserviceapp.CreateServiceInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		ps, err := svc.Create(r.Context(), projectID, input)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, ps)
	}
}

func handleGetService(svc *projectserviceapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID, err := requireParam(r, "serviceId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		ps, err := svc.GetByID(r.Context(), serviceID)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, ps)
	}
}
