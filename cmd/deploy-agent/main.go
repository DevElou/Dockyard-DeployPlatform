package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := getEnv("DOCKYARD_AGENT_ADDR", ":8090")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"dockyard-deploy-agent"}`))
	})

	log.Printf("dockyard deploy-agent listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
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
