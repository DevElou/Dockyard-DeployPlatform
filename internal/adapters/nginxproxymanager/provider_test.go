package nginxproxymanager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elouan/dockyard/internal/ports/routing"
)

type fakeNPM struct {
	hosts []proxyHost
	next  int
}

func newFakeNPM() *fakeNPM { return &fakeNPM{next: 1} }

func (f *fakeNPM) handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/tokens", func(w http.ResponseWriter, r *http.Request) {
		expires := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
		_ = json.NewEncoder(w).Encode(tokenResponse{Token: "test-token", Expires: expires})
	})

	mux.HandleFunc("GET /api/proxy-hosts", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(f.hosts)
	})

	mux.HandleFunc("POST /api/proxy-hosts", func(w http.ResponseWriter, r *http.Request) {
		var h proxyHost
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.ID = f.next
		f.next++
		f.hosts = append(f.hosts, h)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(h)
	})

	mux.HandleFunc("PUT /api/proxy-hosts/{id}", func(w http.ResponseWriter, r *http.Request) {
		var updated proxyHost
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for i, h := range f.hosts {
			if h.ID == updated.ID {
				f.hosts[i] = updated
				_ = json.NewEncoder(w).Encode(updated)
				return
			}
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("DELETE /api/proxy-hosts/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		var id int
		_, _ = parseID(idStr, &id)
		for i, h := range f.hosts {
			if h.ID == id {
				f.hosts = append(f.hosts[:i], f.hosts[i+1:]...)
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		http.NotFound(w, r)
	})

	return mux
}

func parseID(s string, out *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	*out = n
	return n, nil
}

func newTestProvider(t *testing.T, srv *httptest.Server) *Provider {
	t.Helper()
	cfg := Config{
		BaseURL:       srv.URL,
		Identity:      "admin",
		Secret:        "pass",
		ForwardScheme: "http",
		Timeout:       5 * time.Second,
	}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	return p
}

func TestProvider_EnsureRoute_Create(t *testing.T) {
	fake := newFakeNPM()
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	p := newTestProvider(t, srv)
	err := p.EnsureRoute(routing.RouteRequest{
		Hostname:    "app.example.com",
		ForwardHost: "192.168.1.10",
		TargetPort:  3000,
		TLS:         false,
	})
	if err != nil {
		t.Fatalf("EnsureRoute: %v", err)
	}

	if len(fake.hosts) != 1 {
		t.Fatalf("expected 1 proxy host, got %d", len(fake.hosts))
	}
	h := fake.hosts[0]
	if len(h.DomainNames) == 0 || h.DomainNames[0] != "app.example.com" {
		t.Errorf("domain_names = %v, want [app.example.com]", h.DomainNames)
	}
	if h.ForwardHost != "192.168.1.10" {
		t.Errorf("forward_host = %q, want 192.168.1.10", h.ForwardHost)
	}
	if h.ForwardPort != 3000 {
		t.Errorf("forward_port = %d, want 3000", h.ForwardPort)
	}
}

func TestProvider_EnsureRoute_Update(t *testing.T) {
	fake := newFakeNPM()
	fake.hosts = []proxyHost{{
		ID:          1,
		DomainNames: []string{"app.example.com"},
		ForwardHost: "192.168.1.10",
		ForwardPort: 3000,
		ForwardScheme: "http",
	}}
	fake.next = 2

	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	p := newTestProvider(t, srv)
	err := p.EnsureRoute(routing.RouteRequest{
		Hostname:    "app.example.com",
		ForwardHost: "192.168.1.10",
		TargetPort:  4000, // changed port
		TLS:         false,
	})
	if err != nil {
		t.Fatalf("EnsureRoute: %v", err)
	}

	if fake.hosts[0].ForwardPort != 4000 {
		t.Errorf("forward_port = %d, want 4000 after update", fake.hosts[0].ForwardPort)
	}
	if len(fake.hosts) != 1 {
		t.Errorf("expected 1 host, got %d (update must not create a duplicate)", len(fake.hosts))
	}
}

func TestProvider_EnsureRoute_NoOp(t *testing.T) {
	var putCalled bool
	fake := newFakeNPM()
	fake.hosts = []proxyHost{{
		ID:            1,
		DomainNames:   []string{"app.example.com"},
		ForwardHost:   "192.168.1.10",
		ForwardPort:   3000,
		ForwardScheme: "http",
		SSLForced:     false,
	}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			putCalled = true
		}
		fake.handler().ServeHTTP(w, r)
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	err := p.EnsureRoute(routing.RouteRequest{
		Hostname:    "app.example.com",
		ForwardHost: "192.168.1.10",
		TargetPort:  3000,
		TLS:         false,
	})
	if err != nil {
		t.Fatalf("EnsureRoute: %v", err)
	}

	if putCalled {
		t.Error("expected no PUT call when settings are identical")
	}
}

func TestProvider_DeleteRoute_Exists(t *testing.T) {
	fake := newFakeNPM()
	fake.hosts = []proxyHost{{ID: 1, DomainNames: []string{"del.example.com"}}}

	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	p := newTestProvider(t, srv)
	if err := p.DeleteRoute("del.example.com"); err != nil {
		t.Fatalf("DeleteRoute: %v", err)
	}

	if len(fake.hosts) != 0 {
		t.Errorf("expected 0 hosts after delete, got %d", len(fake.hosts))
	}
}

func TestProvider_DeleteRoute_NotFound(t *testing.T) {
	fake := newFakeNPM()
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	p := newTestProvider(t, srv)
	// Should return nil when hostname doesn't exist
	if err := p.DeleteRoute("unknown.example.com"); err != nil {
		t.Fatalf("DeleteRoute on unknown host: %v", err)
	}
}

func TestProvider_EnsureRoute_NPMError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tokens" {
			expires := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
			_ = json.NewEncoder(w).Encode(tokenResponse{Token: "tok", Expires: expires})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	err := p.EnsureRoute(routing.RouteRequest{Hostname: "app.example.com", ForwardHost: "host", TargetPort: 80})
	if err == nil {
		t.Fatal("expected error on NPM 500, got nil")
	}
}
