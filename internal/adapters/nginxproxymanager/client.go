package nginxproxymanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type npmClient struct {
	baseURL    string
	identity   string
	secret     string
	httpClient *http.Client

	mu             sync.Mutex
	cachedToken    string
	tokenExpiresAt time.Time
}

func newClient(cfg Config) *npmClient {
	return &npmClient{
		baseURL:  cfg.BaseURL,
		identity: cfg.Identity,
		secret:   cfg.Secret,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (c *npmClient) token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedToken != "" && time.Until(c.tokenExpiresAt) > 60*time.Second {
		return c.cachedToken, nil
	}
	return c.fetchToken(ctx)
}

func (c *npmClient) fetchToken(ctx context.Context) (string, error) {
	body, err := json.Marshal(tokenRequest{Identity: c.identity, Secret: c.secret})
	if err != nil {
		return "", fmt.Errorf("npm: marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/tokens", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("npm: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("npm: token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("npm: auth failed (status %d)", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("npm: decode token response: %w", err)
	}

	expires := parseExpiry(tr.Expires)
	c.cachedToken = tr.Token
	c.tokenExpiresAt = expires
	return c.cachedToken, nil
}

// parseExpiry parses the NPM token expiry string, falling back to 1 hour from now.
func parseExpiry(raw string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return time.Now().Add(time.Hour)
}

func (c *npmClient) invalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachedToken = ""
	c.tokenExpiresAt = time.Time{}
}

// do performs an authenticated HTTP request, retrying once on 401.
func (c *npmClient) do(ctx context.Context, method, path string, bodyObj any) (*http.Response, error) {
	resp, err := c.doOnce(ctx, method, path, bodyObj)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		c.invalidateToken()
		return c.doOnce(ctx, method, path, bodyObj)
	}
	return resp, nil
}

func (c *npmClient) doOnce(ctx context.Context, method, path string, bodyObj any) (*http.Response, error) {
	tok, err := c.token(ctx)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if bodyObj != nil {
		b, err := json.Marshal(bodyObj)
		if err != nil {
			return nil, fmt.Errorf("npm: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("npm: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if bodyObj != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("npm: %s %s: %w", method, path, err)
	}
	return resp, nil
}
