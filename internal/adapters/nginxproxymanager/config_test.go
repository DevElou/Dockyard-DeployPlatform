package nginxproxymanager

import (
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantEnabled bool
		wantErr     bool
		wantScheme  string
		wantTimeout time.Duration
	}{
		{
			name:        "no URL disables NPM",
			env:         map[string]string{},
			wantEnabled: false,
		},
		{
			name:        "URL without identity returns error",
			env:         map[string]string{"DOCKYARD_NPM_URL": "http://npm.local:81"},
			wantEnabled: false,
			wantErr:     true,
		},
		{
			name: "URL without secret returns error",
			env: map[string]string{
				"DOCKYARD_NPM_URL":      "http://npm.local:81",
				"DOCKYARD_NPM_IDENTITY": "admin@example.com",
			},
			wantEnabled: false,
			wantErr:     true,
		},
		{
			name: "full config enabled with defaults",
			env: map[string]string{
				"DOCKYARD_NPM_URL":      "http://npm.local:81",
				"DOCKYARD_NPM_IDENTITY": "admin@example.com",
				"DOCKYARD_NPM_SECRET":   "password",
			},
			wantEnabled: true,
			wantScheme:  "http",
			wantTimeout: 15 * time.Second,
		},
		{
			name: "custom scheme and timeout",
			env: map[string]string{
				"DOCKYARD_NPM_URL":                    "http://npm.local:81",
				"DOCKYARD_NPM_IDENTITY":               "admin@example.com",
				"DOCKYARD_NPM_SECRET":                 "password",
				"DOCKYARD_NPM_DEFAULT_FORWARD_SCHEME": "https",
				"DOCKYARD_NPM_REQUEST_TIMEOUT":        "30s",
			},
			wantEnabled: true,
			wantScheme:  "https",
			wantTimeout: 30 * time.Second,
		},
		{
			name: "invalid timeout returns error",
			env: map[string]string{
				"DOCKYARD_NPM_URL":             "http://npm.local:81",
				"DOCKYARD_NPM_IDENTITY":        "admin@example.com",
				"DOCKYARD_NPM_SECRET":          "password",
				"DOCKYARD_NPM_REQUEST_TIMEOUT": "not-a-duration",
			},
			wantEnabled: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, enabled, err := LoadConfigFromEnv()

			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if enabled != tt.wantEnabled {
				t.Fatalf("LoadConfigFromEnv() enabled = %v, want %v", enabled, tt.wantEnabled)
			}
			if !tt.wantEnabled || tt.wantErr {
				return
			}

			if cfg.ForwardScheme != tt.wantScheme {
				t.Errorf("ForwardScheme = %q, want %q", cfg.ForwardScheme, tt.wantScheme)
			}
			if cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", cfg.Timeout, tt.wantTimeout)
			}
		})
	}
}
