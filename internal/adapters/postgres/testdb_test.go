//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DOCKYARD_TEST_DSN")
	if dsn == "" {
		t.Skip("DOCKYARD_TEST_DSN not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open test pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("ping test db: %v", err)
	}

	t.Cleanup(func() {
		// truncate in FK-safe order
		_, _ = pool.Exec(context.Background(), `
			TRUNCATE operation_events, deployment_steps, deployments, build_jobs, releases,
			         domains, project_services, environment_variables, environment_sets,
			         project_runtime_targets, projects, runtime_targets CASCADE
		`)
		pool.Close()
	})

	return pool
}
