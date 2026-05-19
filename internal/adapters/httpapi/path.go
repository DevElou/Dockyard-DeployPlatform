package httpapi

import (
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
	switch err {
	case postgres.ErrProjectNotFound,
		postgres.ErrRuntimeTargetNotFound,
		postgres.ErrReleaseNotFound,
		postgres.ErrDeploymentNotFound,
		postgres.ErrDomainNotFound:
		httpjson.Error(w, http.StatusNotFound, "not_found", err.Error())
	case postgres.ErrProjectSlugExists,
		postgres.ErrRuntimeTargetSlugExists,
		postgres.ErrReleaseVersionExists,
		postgres.ErrReleaseDigestExists,
		postgres.ErrDomainHostnameExists,
		postgres.ErrProjectRuntimeTargetExists:
		httpjson.Error(w, http.StatusConflict, "conflict", err.Error())
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}
