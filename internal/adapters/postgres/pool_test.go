package postgres

import (
	"testing"
)

func TestLoadConfigFromEnv_MissingDSN(t *testing.T) {
	t.Setenv("DOCKYARD_DATABASE_URL", "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error when DOCKYARD_DATABASE_URL is empty")
	}
}

func TestLoadConfigFromEnv_Present(t *testing.T) {
	t.Setenv("DOCKYARD_DATABASE_URL", "postgresql://root@localhost:26257/dockyard?sslmode=disable")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DSN == "" {
		t.Fatal("expected non-empty DSN")
	}
}
