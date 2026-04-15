package httpapi

import (
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	httpjson.Write(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "dockyard-control-plane-api",
	})
}
