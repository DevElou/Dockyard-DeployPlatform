package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elouan/dockyard/internal/adapters/github"
	"github.com/elouan/dockyard/internal/adapters/httpapi"
	"github.com/elouan/dockyard/internal/adapters/nginxproxymanager"
	"github.com/elouan/dockyard/internal/adapters/postgres"
	deploymentapp "github.com/elouan/dockyard/internal/application/deployment"
	domainsvc "github.com/elouan/dockyard/internal/application/domainsvc"
	envapp "github.com/elouan/dockyard/internal/application/environment"
	projectapp "github.com/elouan/dockyard/internal/application/project"
	projectserviceapp "github.com/elouan/dockyard/internal/application/projectservice"
	releaseapp "github.com/elouan/dockyard/internal/application/release"
	runtimetargetapp "github.com/elouan/dockyard/internal/application/runtimetarget"
	"github.com/elouan/dockyard/internal/ports/routing"
)

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
	projectRepo := postgres.NewProjectRepository(pool)
	src := github.NewSourceProvider(githubToken, projectRepo)

	router := httpapi.NewRouter(httpapi.RouterDeps{
		ProjectService:        projectapp.NewService(projectRepo),
		RuntimeTargetService:  runtimetargetapp.NewService(postgres.NewRuntimeTargetRepository(pool)),
		ReleaseService:        releaseapp.NewService(postgres.NewReleaseRepository(pool), src),
		DeploymentService:     deploymentapp.NewService(postgres.NewDeploymentRepository(pool)),
		DomainService:         domainsvc.NewService(postgres.NewDomainRepository(pool), routingProvider),
		ProjectServiceService: projectserviceapp.NewService(postgres.NewProjectServiceRepository(pool)),
		EnvironmentService:    envapp.NewService(postgres.NewEnvironmentSetRepository(pool), postgres.NewEnvironmentVariableRepository(pool)),
	})

	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.Printf("dockyard control-plane-api listening on %s", addr)
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
