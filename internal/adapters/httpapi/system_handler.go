package httpapi

import (
	"net/http"

	"github.com/elouan/dockyard/internal/adapters/httpjson"
)

type IntegrationInfo struct {
	Enabled bool   `json:"enabled"`
	BaseURL string `json:"baseUrl,omitempty"`
}

type SystemIntegrations struct {
	GitHub   IntegrationInfo `json:"github"`
	NPM      IntegrationInfo `json:"npm"`
	DNS      IntegrationInfo `json:"dns"`
	Registry IntegrationInfo `json:"registry"`
}

type SystemInfo struct {
	Version      string             `json:"version"`
	Integrations SystemIntegrations `json:"integrations"`
}

func handleSystemInfo(info SystemInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpjson.Write(w, http.StatusOK, info)
	}
}
