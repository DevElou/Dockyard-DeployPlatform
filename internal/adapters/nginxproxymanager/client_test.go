package nginxproxymanager

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func makeTokenServer(t *testing.T, token, expires string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tokens" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(tokenResponse{Token: token, Expires: expires})
	}))
}

func TestClient_TokenFetch(t *testing.T) {
	expires := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	srv := makeTokenServer(t, "tok-abc", expires, http.StatusOK)
	defer srv.Close()

	c := newClient(Config{BaseURL: srv.URL, Identity: "user", Secret: "pass", Timeout: 5 * time.Second})

	tok, err := c.token(context.Background())
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if tok != "tok-abc" {
		t.Errorf("token = %q, want tok-abc", tok)
	}
}

func TestClient_TokenCacheHit(t *testing.T) {
	expires := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_ = json.NewEncoder(w).Encode(tokenResponse{Token: "tok", Expires: expires})
	}))
	defer srv.Close()

	c := newClient(Config{BaseURL: srv.URL, Identity: "u", Secret: "p", Timeout: 5 * time.Second})
	ctx := context.Background()

	if _, err := c.token(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.token(ctx); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 {
		t.Errorf("expected 1 token fetch, got %d", calls.Load())
	}
}

func TestClient_TokenRefreshAfterExpiry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		expires := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
		_ = json.NewEncoder(w).Encode(tokenResponse{Token: "tok", Expires: expires})
	}))
	defer srv.Close()

	c := newClient(Config{BaseURL: srv.URL, Identity: "u", Secret: "p", Timeout: 5 * time.Second})
	ctx := context.Background()

	// Manually set an expired token
	c.mu.Lock()
	c.cachedToken = "old-tok"
	c.tokenExpiresAt = time.Now().Add(30 * time.Second) // within 60s buffer
	c.mu.Unlock()

	if _, err := c.token(ctx); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 {
		t.Errorf("expected token refresh, got %d calls", calls.Load())
	}
}

func TestClient_Do_Retry401(t *testing.T) {
	var tokenCalls atomic.Int32
	var apiCalls atomic.Int32
	returnUnauthorized := true

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tokens" {
			tokenCalls.Add(1)
			expires := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
			_ = json.NewEncoder(w).Encode(tokenResponse{Token: "tok", Expires: expires})
			return
		}
		apiCalls.Add(1)
		if returnUnauthorized {
			returnUnauthorized = false
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	c := newClient(Config{BaseURL: srv.URL, Identity: "u", Secret: "p", Timeout: 5 * time.Second})
	ctx := context.Background()

	resp, err := c.do(ctx, http.MethodGet, "/api/proxy-hosts", nil)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if apiCalls.Load() != 2 {
		t.Errorf("expected 2 API calls (first 401 + retry), got %d", apiCalls.Load())
	}
	if tokenCalls.Load() != 2 {
		t.Errorf("expected 2 token fetches (initial + after 401), got %d", tokenCalls.Load())
	}
}

func TestClient_AuthFailure(t *testing.T) {
	srv := makeTokenServer(t, "", "", http.StatusUnauthorized)
	defer srv.Close()

	c := newClient(Config{BaseURL: srv.URL, Identity: "u", Secret: "bad", Timeout: 5 * time.Second})

	_, err := c.token(context.Background())
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
}

func TestParseExpiry(t *testing.T) {
	tests := []struct {
		raw     string
		wantErr bool
	}{
		{"2026-06-01T12:00:00Z", false},
		{"2026-06-01T12:00:00.000Z", false},
		{"2026-06-01T12:00:00+02:00", false},
		{"not-a-date", true}, // falls back to time.Now + 1h, not an error
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			result := parseExpiry(tt.raw)
			if result.IsZero() {
				t.Error("expected non-zero time")
			}
			// For invalid input, result should be approximately now + 1h
			if tt.wantErr {
				diff := time.Until(result)
				if diff < 59*time.Minute || diff > 61*time.Minute {
					t.Errorf("fallback expiry %v not near 1h from now", result)
				}
			}
		})
	}
}
