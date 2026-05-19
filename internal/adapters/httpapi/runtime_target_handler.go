package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	runtimetargetapp "github.com/elouan/dockyard/internal/application/runtimetarget"
)

func handleListRuntimeTargets(svc *runtimetargetapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targets, err := svc.List(r.Context())
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_runtime_targets_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": targets})
	}
}

func handleCreateRuntimeTarget(svc *runtimetargetapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input runtimetargetapp.CreateRuntimeTargetInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		target, err := svc.Create(r.Context(), input)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, target)
	}
}

func handleGetRuntimeTarget(svc *runtimetargetapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		target, err := svc.GetByID(r.Context(), id)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, target)
	}
}

func handleEnableRuntimeTarget(svc *runtimetargetapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		if err := svc.Enable(r.Context(), id); err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}

func handleDisableRuntimeTarget(svc *runtimetargetapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "id")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		if err := svc.Disable(r.Context(), id); err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusNoContent, nil)
	}
}
