package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elouan/dockyard/internal/adapters/dockerregistry"
	"github.com/elouan/dockyard/internal/adapters/github"
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

	agentKey := getEnv("DOCKYARD_AGENT_KEY", "")

	factory := func(target domain.RuntimeTarget) (agent.Client, error) {
		if target.Endpoint == "" {
			return nil, fmt.Errorf("runtime target %s has no endpoint", target.ID)
		}
		return httpclient.NewAgentClient(target.Endpoint, agentKey), nil
	}

	projectRepo := postgres.NewProjectRepository(pool)
	src := github.NewSourceProvider(getEnv("DOCKYARD_GITHUB_TOKEN", ""), projectRepo)
	builder := dockerregistry.NewBuilder(getEnv("DOCKYARD_REGISTRY_URL", ""))

	buildWorker := NewBuildWorker(
		postgres.NewReleaseRepository(pool),
		projectRepo,
		src,
		builder,
	)

	deployWorker := NewDeployWorker(
		postgres.NewDeploymentRepository(pool),
		postgres.NewReleaseRepository(pool),
		projectRepo,
		postgres.NewRuntimeTargetRepository(pool),
		postgres.NewProjectServiceRepository(pool),
		postgres.NewEnvironmentSetRepository(pool),
		postgres.NewEnvironmentVariableRepository(pool),
		factory,
	)

	buildDone := make(chan struct{})
	go func() {
		defer close(buildDone)
		buildWorker.Run(ctx)
	}()

	deployWorker.Run(ctx)
	<-buildDone
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
