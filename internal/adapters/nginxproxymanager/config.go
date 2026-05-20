package nginxproxymanager

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	BaseURL       string
	Identity      string
	Secret        string
	ForwardScheme string
	Timeout       time.Duration
}

// LoadConfigFromEnv reads NPM configuration from environment variables.
// Returns (config, enabled, error). enabled is false when DOCKYARD_NPM_URL is empty.
func LoadConfigFromEnv() (Config, bool, error) {
	baseURL := os.Getenv("DOCKYARD_NPM_URL")
	if baseURL == "" {
		return Config{}, false, nil
	}

	identity := os.Getenv("DOCKYARD_NPM_IDENTITY")
	if identity == "" {
		return Config{}, false, fmt.Errorf("DOCKYARD_NPM_IDENTITY is required when DOCKYARD_NPM_URL is set")
	}

	secret := os.Getenv("DOCKYARD_NPM_SECRET")
	if secret == "" {
		return Config{}, false, fmt.Errorf("DOCKYARD_NPM_SECRET is required when DOCKYARD_NPM_URL is set")
	}

	scheme := os.Getenv("DOCKYARD_NPM_DEFAULT_FORWARD_SCHEME")
	if scheme == "" {
		scheme = "http"
	}

	timeout := 15 * time.Second
	if raw := os.Getenv("DOCKYARD_NPM_REQUEST_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, false, fmt.Errorf("invalid DOCKYARD_NPM_REQUEST_TIMEOUT: %w", err)
		}
		timeout = d
	}

	return Config{
		BaseURL:       baseURL,
		Identity:      identity,
		Secret:        secret,
		ForwardScheme: scheme,
		Timeout:       timeout,
	}, true, nil
}
