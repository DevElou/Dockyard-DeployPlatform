package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
	"github.com/elouan/dockyard/internal/adapters/postgres"
)

func requireParam(r *http.Request, name string) (string, error) {
	v := r.PathValue(name)
	if v == "" {
		return "", fmt.Errorf("missing path parameter: %s", name)
	}
	return v, nil
}

func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, postgres.ErrProjectNotFound),
		errors.Is(err, postgres.ErrRuntimeTargetNotFound),
		errors.Is(err, postgres.ErrReleaseNotFound),
		errors.Is(err, postgres.ErrDeploymentNotFound),
		errors.Is(err, postgres.ErrDomainNotFound),
		errors.Is(err, postgres.ErrProjectServiceNotFound),
		errors.Is(err, postgres.ErrEnvironmentSetNotFound),
		errors.Is(err, postgres.ErrEnvironmentVariableNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, postgres.ErrProjectSlugExists),
		errors.Is(err, postgres.ErrRuntimeTargetSlugExists),
		errors.Is(err, postgres.ErrReleaseVersionExists),
		errors.Is(err, postgres.ErrReleaseDigestExists),
		errors.Is(err, postgres.ErrDomainHostnameExists),
		errors.Is(err, postgres.ErrProjectRuntimeTargetExists),
		errors.Is(err, postgres.ErrProjectServiceNameExists),
		errors.Is(err, postgres.ErrEnvironmentSetNameExists):
		httpjson.Error(w, http.StatusConflict, "conflict", err.Error())
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

