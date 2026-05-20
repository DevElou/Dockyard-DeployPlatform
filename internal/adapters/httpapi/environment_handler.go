package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	envapp "github.com/elouan/dockyard/internal/application/environment"
)

func handleListEnvironmentSets(svc *envapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		sets, err := svc.ListSets(r.Context(), projectID)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_environments_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": sets})
	}
}

func handleCreateEnvironmentSet(svc *envapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input envapp.CreateSetInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		set, err := svc.CreateSet(r.Context(), projectID, input)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, set)
	}
}

func handleListVariables(svc *envapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envID, err := requireParam(r, "envId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		vars, err := svc.ListVariables(r.Context(), envID)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_variables_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": vars})
	}
}

func handleUpsertVariable(svc *envapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envID, err := requireParam(r, "envId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input envapp.UpsertVariableInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		if err := svc.UpsertVariable(r.Context(), envID, input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "upsert_variable_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}

func handleDeleteVariable(svc *envapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		varID, err := requireParam(r, "varId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		if err := svc.DeleteVariable(r.Context(), varID); err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}
