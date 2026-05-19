# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Local dev infrastructure (CockroachDB, Redis, Registry)
make local-infra-up
make local-infra-down
make local-infra-logs

# Run Go services
make run-api       # control-plane-api
make run-worker    # orchestrator-worker
make run-agent     # deploy-agent

# Go build, format, test
make build
make fmt
make test
go test ./internal/application/project/...   # single package

# Frontend
cd web && npm install
make web-dev       # or: cd web && npm run dev
cd web && npm run build && npm run lint
```

## Architecture

Dockyard is a private deployment platform (Vercel-like) targeting Docker hosts on a homelab ESXi infrastructure.

### Three Go binaries

| Binary | Path | Role |
|---|---|---|
| `control-plane-api` | `cmd/control-plane-api` | HTTP API, state owner, publishes async jobs |
| `orchestrator-worker` | `cmd/orchestrator-worker` | Async workflows: build, release, deploy, rollback |
| `deploy-agent` | `cmd/deploy-agent` | Runs on each Docker host, executes deployment specs locally |

### Layered architecture (hexagonal)

```
cmd/*/main.go          ← wiring only: config, instantiate adapters, start server
internal/domain/       ← pure business types and invariants (no external SDK imports)
internal/application/  ← use cases, orchestrate domain via port interfaces
internal/ports/        ← interfaces: repository, runtime, source, dns, routing, queue, registry, agent
internal/adapters/     ← concrete implementations (HTTP router, in-memory repo, etc.)
```

Request flow: `HTTP → adapters/httpapi → application → ports → adapters/concrete`

### Port interfaces (V1)

Defined in `internal/ports/`, each interface isolates one infrastructure concern:
- `repository.ProjectRepository` — canonical state (CockroachDB target, currently `adapters/memory`)
- `runtime.Driver` — Docker container lifecycle
- `source.Provider` — GitHub repo access
- `registry.Builder` — Docker image build + push
- `agent.Client` — send `DeploymentSpec` to a deploy-agent
- `dns.Provider` — DuckDNS
- `routing.Provider` — Nginx Proxy Manager

### Core domain resources

`Project → Release → Deployment`, plus `RuntimeTarget` and `Domain`.

- `Release` is **immutable** after creation (image digest + tag locked)
- `Deployment` is a business event; rollback creates a new deployment, never rewrites history
- The deploy-agent **executes only** — no global orchestration logic

### Current state

The scaffold is functional but pre-persistence: `adapters/memory` is the active repository. CockroachDB adapters, GitHub integration, and Docker runtime driver are not yet implemented. The recommended next vertical slice: SQL migrations → repository adapters → project/release/deployment CRUD → real deploy-agent → Nginx Proxy Manager routing.

## Module

```
module github.com/elouan/dockyard   (go 1.24)
```
