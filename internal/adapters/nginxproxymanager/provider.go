package nginxproxymanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elouan/dockyard/internal/ports/routing"
)

type Provider struct {
	client        *npmClient
	defaultScheme string
	timeout       time.Duration
}

func NewProvider(cfg Config) (*Provider, error) {
	return &Provider{
		client:        newClient(cfg),
		defaultScheme: cfg.ForwardScheme,
		timeout:       cfg.Timeout,
	}, nil
}

func (p *Provider) EnsureRoute(req routing.RouteRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	existing, err := p.findByHostname(ctx, req.Hostname)
	if err != nil {
		return fmt.Errorf("npm: ensure route %s: %w", req.Hostname, err)
	}

	scheme := req.ForwardScheme
	if scheme == "" {
		scheme = p.defaultScheme
	}

	desired := proxyHost{
		DomainNames:    []string{req.Hostname},
		ForwardHost:    req.ForwardHost,
		ForwardPort:    req.TargetPort,
		ForwardScheme:  scheme,
		SSLForced:      req.TLS,
		CachingEnabled: false,
		BlockExploits:  true,
	}

	if existing == nil {
		return p.create(ctx, desired)
	}

	if existing.ForwardHost == desired.ForwardHost &&
		existing.ForwardPort == desired.ForwardPort &&
		existing.ForwardScheme == desired.ForwardScheme &&
		existing.SSLForced == desired.SSLForced {
		return nil
	}

	desired.ID = existing.ID
	return p.update(ctx, desired)
}

func (p *Provider) DeleteRoute(hostname string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	existing, err := p.findByHostname(ctx, hostname)
	if err != nil {
		return fmt.Errorf("npm: delete route %s: %w", hostname, err)
	}
	if existing == nil {
		return nil
	}

	resp, err := p.client.do(ctx, http.MethodDelete, fmt.Sprintf("/api/proxy-hosts/%d", existing.ID), nil)
	if err != nil {
		return fmt.Errorf("npm: delete proxy host %d: %w", existing.ID, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("npm: delete proxy host %d: unexpected status %d", existing.ID, resp.StatusCode)
	}
	return nil
}

func (p *Provider) findByHostname(ctx context.Context, hostname string) (*proxyHost, error) {
	resp, err := p.client.do(ctx, http.MethodGet, "/api/proxy-hosts", nil)
	if err != nil {
		return nil, fmt.Errorf("list proxy hosts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list proxy hosts: status %d: %s", resp.StatusCode, string(body))
	}

	var hosts []proxyHost
	if err := json.NewDecoder(resp.Body).Decode(&hosts); err != nil {
		return nil, fmt.Errorf("decode proxy hosts: %w", err)
	}

	for i := range hosts {
		for _, dn := range hosts[i].DomainNames {
			if dn == hostname {
				return &hosts[i], nil
			}
		}
	}
	return nil, nil
}

func (p *Provider) create(ctx context.Context, h proxyHost) error {
	resp, err := p.client.do(ctx, http.MethodPost, "/api/proxy-hosts", h)
	if err != nil {
		return fmt.Errorf("create proxy host: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusConflict {
		return nil
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create proxy host: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (p *Provider) update(ctx context.Context, h proxyHost) error {
	resp, err := p.client.do(ctx, http.MethodPut, fmt.Sprintf("/api/proxy-hosts/%d", h.ID), h)
	if err != nil {
		return fmt.Errorf("update proxy host %d: %w", h.ID, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update proxy host %d: unexpected status %d", h.ID, resp.StatusCode)
	}
	return nil
}
