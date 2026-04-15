package main

import (
	"log"
	"net/http"
	"os"

	"github.com/elouan/dockyard/internal/adapters/httpapi"
	"github.com/elouan/dockyard/internal/adapters/memory"
	projectapp "github.com/elouan/dockyard/internal/application/project"
)

func main() {
	addr := getEnv("DOCKYARD_API_ADDR", ":8080")

	projectRepository := memory.NewProjectRepository()
	projectService := projectapp.NewService(projectRepository)

	router := httpapi.NewRouter(httpapi.RouterDeps{
		ProjectService: projectService,
	})

	log.Printf("dockyard control-plane-api listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
