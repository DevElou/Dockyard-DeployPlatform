package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	domainsvc "github.com/elouan/dockyard/internal/application/domainsvc"
)

func handleListDomains(svc *domainsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		domains, err := svc.List(r.Context(), projectID)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_domains_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": domains})
	}
}

func handleCreateDomain(svc *domainsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input domainsvc.CreateDomainInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		d, err := svc.Create(r.Context(), projectID, input)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, d)
	}
}

func handleGetDomain(svc *domainsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "domainId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		d, err := svc.GetByID(r.Context(), id)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, d)
	}
}

func handleDeleteDomain(svc *domainsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "domainId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		if err := svc.Delete(r.Context(), id); err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}
