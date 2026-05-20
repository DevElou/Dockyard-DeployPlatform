package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elouan/dockyard/internal/adapters/httpclient"
	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgCfg, err := postgres.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, pgCfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	factory := func(target domain.RuntimeTarget) (agent.Client, error) {
		if target.Endpoint == "" {
			return nil, fmt.Errorf("runtime target %s has no endpoint", target.ID)
		}
		agentKey := getEnv("DOCKYARD_AGENT_KEY", "")
		return httpclient.NewAgentClient(target.Endpoint, agentKey), nil
	}

	worker := NewWorker(
		postgres.NewDeploymentRepository(pool),
		postgres.NewReleaseRepository(pool),
		postgres.NewProjectRepository(pool),
		postgres.NewRuntimeTargetRepository(pool),
		factory,
	)

	worker.Run(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
