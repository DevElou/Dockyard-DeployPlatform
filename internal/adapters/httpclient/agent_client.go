package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elouan/dockyard/internal/ports/agent"
)

type AgentClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewAgentClient(baseURL, apiKey string) *AgentClient {
	return &AgentClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *AgentClient) Deploy(ctx context.Context, req agent.DeployRequest) (agent.DeployResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return agent.DeployResponse{}, fmt.Errorf("agent client: marshal deploy request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/deployments", bytes.NewReader(body))
	if err != nil {
		return agent.DeployResponse{}, fmt.Errorf("agent client: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Agent-Key", c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return agent.DeployResponse{}, fmt.Errorf("agent client: deploy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return agent.DeployResponse{}, fmt.Errorf("agent client: deploy: unexpected status %d", resp.StatusCode)
	}

	var result agent.DeployResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return agent.DeployResponse{}, fmt.Errorf("agent client: decode deploy response: %w", err)
	}
	return result, nil
}

func (c *AgentClient) GetStatus(ctx context.Context, deploymentID string) (agent.StatusResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/deployments/"+deploymentID, nil)
	if err != nil {
		return agent.StatusResponse{}, fmt.Errorf("agent client: build request: %w", err)
	}
	httpReq.Header.Set("X-Agent-Key", c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return agent.StatusResponse{}, fmt.Errorf("agent client: get status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return agent.StatusResponse{}, fmt.Errorf("agent client: deployment %s not found", deploymentID)
	}
	if resp.StatusCode != http.StatusOK {
		return agent.StatusResponse{}, fmt.Errorf("agent client: get status: unexpected status %d", resp.StatusCode)
	}

	var result agent.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return agent.StatusResponse{}, fmt.Errorf("agent client: decode status response: %w", err)
	}
	return result, nil
}

func (c *AgentClient) Remove(ctx context.Context, deploymentID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.baseURL+"/deployments/"+deploymentID, nil)
	if err != nil {
		return fmt.Errorf("agent client: build request: %w", err)
	}
	httpReq.Header.Set("X-Agent-Key", c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("agent client: remove: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent client: remove: unexpected status %d", resp.StatusCode)
	}
	return nil
}
