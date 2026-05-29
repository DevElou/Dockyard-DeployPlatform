package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	releaseapp "github.com/elouan/dockyard/internal/application/release"
)

func handleListReleases(svc *releaseapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		releases, err := svc.List(r.Context(), projectID)
		if err != nil {
			httpjson.Error(w, http.StatusInternalServerError, "list_releases_failed", err.Error())
			return
		}
		httpjson.Write(w, http.StatusOK, map[string]any{"items": releases})
	}
}

func handleCreateRelease(svc *releaseapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requireParam(r, "projectId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		var input releaseapp.CreateReleaseInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
			return
		}

		release, err := svc.Create(r.Context(), projectID, input)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusCreated, release)
	}
}

func handleGetRelease(svc *releaseapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "releaseId")
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "missing_param", err.Error())
			return
		}

		release, err := svc.GetByID(r.Context(), id)
		if err != nil {
			mapError(w, err)
			return
		}
		httpjson.Write(w, http.StatusOK, release)
	}
}

func handleListReleaseEvents(svc *releaseapp.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requireParam(r, "releaseId")
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
