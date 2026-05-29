package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := getEnv("DOCKYARD_AGENT_ADDR", ":8090")
	apiKey := getEnv("DOCKYARD_AGENT_KEY", "")

	if apiKey == "" {
		log.Fatal("DOCKYARD_AGENT_KEY must be set; refusing to start without authentication")
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	h := newAgentHandler(apiKey, shutdownCtx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"dockyard-deploy-agent"}`))
	})
	mux.HandleFunc("POST /deployments", h.authMiddleware(h.handleDeploy))
	mux.HandleFunc("GET /deployments/{id}", h.authMiddleware(h.handleGetStatus))
	mux.HandleFunc("GET /deployments/{id}/logs", h.authMiddleware(h.handleGetLogs))
	mux.HandleFunc("DELETE /deployments/{id}", h.authMiddleware(h.handleRemove))

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("dockyard deploy-agent listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-shutdownCtx.Done()
	stop()

	// Give in-flight requests 30 seconds to finish, then drain background goroutines.
	shutdownDeadline, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownDeadline); err != nil {
		log.Printf("agent: server shutdown: %v", err)
	}

	// Wait for in-flight deployments (they share the shutdownCtx, so they will cancel).
	h.drain()
	log.Println("dockyard deploy-agent stopped")
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
