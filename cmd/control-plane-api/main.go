package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elouan/dockyard/internal/adapters/github"
	"github.com/elouan/dockyard/internal/adapters/httpapi"
	"github.com/elouan/dockyard/internal/adapters/httpclient"
	"github.com/elouan/dockyard/internal/adapters/nginxproxymanager"
	"github.com/elouan/dockyard/internal/adapters/postgres"
	"github.com/elouan/dockyard/internal/application/containerlogs"
	deploymentapp "github.com/elouan/dockyard/internal/application/deployment"
	domainsvc "github.com/elouan/dockyard/internal/application/domainsvc"
	envapp "github.com/elouan/dockyard/internal/application/environment"
	"github.com/elouan/dockyard/internal/application/operationlog"
	projectapp "github.com/elouan/dockyard/internal/application/project"
	projectserviceapp "github.com/elouan/dockyard/internal/application/projectservice"
	releaseapp "github.com/elouan/dockyard/internal/application/release"
	runtimetargetapp "github.com/elouan/dockyard/internal/application/runtimetarget"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
	"github.com/elouan/dockyard/internal/ports/routing"
)

// version is set at build time via -ldflags "-X main.version=x.y.z".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := getEnv("DOCKYARD_API_ADDR", ":8080")

	pgCfg, err := postgres.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, pgCfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	npmCfg, npmEnabled, err := nginxproxymanager.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("npm config: %v", err)
	}
	var routingProvider routing.Provider
	if npmEnabled {
		p, err := nginxproxymanager.NewProvider(npmCfg)
		if err != nil {
			log.Fatalf("npm provider: %v", err)
		}
		routingProvider = p
		log.Printf("npm routing enabled: %s", npmCfg.BaseURL)
	} else {
		routingProvider = &nginxproxymanager.NoopProvider{}
	}

	githubToken := mustEnv("DOCKYARD_GITHUB_TOKEN")
	registryURL := os.Getenv("DOCKYARD_REGISTRY_URL")
	agentKey := os.Getenv("DOCKYARD_AGENT_KEY")

	projectRepo := postgres.NewProjectRepository(pool)
	src := github.NewSourceProvider(githubToken, projectRepo)
	events := operationlog.NewService(postgres.NewOperationLogRepository(pool))

	deploymentRepo := postgres.NewDeploymentRepository(pool)
	runtimeTargetRepo := postgres.NewRuntimeTargetRepository(pool)

	var containerLogsService *containerlogs.Service
	if agentKey != "" {
		factory := func(target domain.RuntimeTarget) (agent.Client, error) {
			if target.Endpoint == "" {
				return nil, fmt.Errorf("runtime target %s has no endpoint", target.ID)
			}
			return httpclient.NewAgentClient(target.Endpoint, agentKey), nil
		}
		containerLogsService = containerlogs.NewService(deploymentRepo, runtimeTargetRepo, factory)
	} else {
		log.Println("container logs disabled (DOCKYARD_AGENT_KEY not set)")
	}

	systemInfo := httpapi.SystemInfo{
		Version: version,
		Integrations: httpapi.SystemIntegrations{
			GitHub:   httpapi.IntegrationInfo{Enabled: githubToken != ""},
			NPM:      httpapi.IntegrationInfo{Enabled: npmEnabled, BaseURL: npmCfg.BaseURL},
			DNS:      httpapi.IntegrationInfo{Enabled: false},
			Registry: httpapi.IntegrationInfo{Enabled: registryURL != "", BaseURL: registryURL},
		},
	}

	router := httpapi.NewRouter(httpapi.RouterDeps{
		ProjectService:        projectapp.NewService(projectRepo),
		RuntimeTargetService:  runtimetargetapp.NewService(runtimeTargetRepo),
		ReleaseService:        releaseapp.NewService(postgres.NewReleaseRepository(pool), src, events),
		DeploymentService:     deploymentapp.NewService(deploymentRepo, events),
		DomainService:         domainsvc.NewService(postgres.NewDomainRepository(pool), routingProvider),
		ProjectServiceService: projectserviceapp.NewService(postgres.NewProjectServiceRepository(pool)),
		EnvironmentService:    envapp.NewService(postgres.NewEnvironmentSetRepository(pool), postgres.NewEnvironmentVariableRepository(pool)),
		ContainerLogsService:  containerLogsService,
		System:                systemInfo,
	})

	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.Printf("dockyard control-plane-api listening on %s (version: %s)", addr, version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func getEnv(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
