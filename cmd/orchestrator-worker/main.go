package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("dockyard orchestrator-worker started")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			log.Println("dockyard orchestrator-worker heartbeat")
		case sig := <-signals:
			log.Printf("dockyard orchestrator-worker shutting down: %s", sig)
			return
		}
	}
}
