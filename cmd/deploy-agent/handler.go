package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/elouan/dockyard/internal/adapters/docker"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/agent"
	"github.com/elouan/dockyard/internal/ports/runtime"
)

type deploymentStore struct {
	mu      sync.RWMutex
	results map[string]runtime.DeploymentResult
}

func newDeploymentStore() *deploymentStore {
	return &deploymentStore{results: make(map[string]runtime.DeploymentResult)}
}

func (s *deploymentStore) set(id string, r runtime.DeploymentResult) {
	s.mu.Lock()
	s.results[id] = r
	s.mu.Unlock()
}

func (s *deploymentStore) get(id string) (runtime.DeploymentResult, bool) {
	s.mu.RLock()
	r, ok := s.results[id]
	s.mu.RUnlock()
	return r, ok
}

func (s *deploymentStore) delete(id string) {
	s.mu.Lock()
	delete(s.results, id)
	s.mu.Unlock()
}

type agentHandler struct {
	driver      *docker.Driver
	store       *deploymentStore
	apiKey      string
	shutdownCtx context.Context
	wg          sync.WaitGroup
}

func newAgentHandler(apiKey string, shutdownCtx context.Context) *agentHandler {
	return &agentHandler{
		driver:      docker.NewDriver(),
		store:       newDeploymentStore(),
		apiKey:      apiKey,
		shutdownCtx: shutdownCtx,
	}
}

// drain waits for all in-flight deployments to finish.
func (h *agentHandler) drain() {
	h.wg.Wait()
}

func (h *agentHandler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Agent-Key") != h.apiKey {
			jsonError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func (h *agentHandler) handleDeploy(w http.ResponseWriter, r *http.Request) {
	var req agent.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	deploymentID := req.Spec.DeploymentID
	if deploymentID == "" {
		jsonError(w, http.StatusBadRequest, "deployment ID is required")
		return
	}

	h.store.set(deploymentID, runtime.DeploymentResult{
		DeploymentID: deploymentID,
		Status:       domain.DeploymentStatusDeploying,
	})

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.executeDeployment(req.Spec)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(agent.DeployResponse{
		Accepted:     true,
		DeploymentID: deploymentID,
	})
}

func (h *agentHandler) executeDeployment(spec runtime.DeploymentSpec) {
	// Derive from shutdownCtx so the goroutine is cancelled on server shutdown.
	ctx, cancel := context.WithTimeout(h.shutdownCtx, 10*time.Minute)
	defer cancel()

	if err := h.driver.PrepareDeployment(ctx, spec); err != nil {
		log.Printf("agent: prepare deployment %s: %v", spec.DeploymentID, err)
		h.store.set(spec.DeploymentID, runtime.DeploymentResult{
			DeploymentID: spec.DeploymentID,
			Status:       domain.DeploymentStatusFailed,
			Message:      err.Error(),
		})
		return
	}

	result, err := h.driver.ApplyRelease(ctx, spec)
	if err != nil {
		log.Printf("agent: apply release %s: %v", spec.DeploymentID, err)
	}
	h.store.set(spec.DeploymentID, result)
}

func (h *agentHandler) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	deploymentID := r.PathValue("id")
	if deploymentID == "" {
		jsonError(w, http.StatusBadRequest, "deployment ID is required")
		return
	}

	result, ok := h.store.get(deploymentID)
	if !ok {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		var err error
		result, err = h.driver.CheckHealth(ctx, deploymentID)
		if err != nil {
			jsonError(w, http.StatusNotFound, "deployment not found")
			return
		}
		h.store.set(deploymentID, result)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agent.StatusResponse{
		DeploymentID: deploymentID,
		Result:       result,
	})
}

func (h *agentHandler) handleRemove(w http.ResponseWriter, r *http.Request) {
	deploymentID := r.PathValue("id")
	if deploymentID == "" {
		jsonError(w, http.StatusBadRequest, "deployment ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := h.driver.DeleteDeployment(ctx, deploymentID); err != nil {
		log.Printf("agent: delete deployment %s: %v", deploymentID, err)
		jsonError(w, http.StatusInternalServerError, "failed to remove deployment")
		return
	}

	h.store.delete(deploymentID)
	w.WriteHeader(http.StatusNoContent)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
