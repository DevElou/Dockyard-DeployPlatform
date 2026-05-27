# Dockyard

Dockyard est une plateforme privée de déploiement inspirée de Vercel, conçue pour piloter une infrastructure homelab ESXi. Elle fournit une interface unique pour connecter un projet GitHub, construire son image Docker, la déployer sur des serveurs Docker et gérer son exposition HTTP(S).

## Architecture générale

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            Opérateur                                    │
│                          (browser / curl)                               │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │ HTTP REST
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       control-plane-api :8080                           │
│                                                                         │
│  POST /projects          POST /releases          POST /deployments      │
│  POST /services          POST /environments      POST /domains          │
└──────────────┬──────────────────────────────────────────────────────────┘
               │ read / write
               ▼
┌──────────────────────────┐
│     CockroachDB           │   (état canonique de toutes les ressources)
└──────────────────────────┘
               ▲
               │ poll toutes les 5-10s
               │
┌──────────────┴───────────────────────────────────────────────────────────┐
│                      orchestrator-worker                                  │
│                                                                           │
│   ┌──────────────────────────┐    ┌─────────────────────────────────┐    │
│   │       BuildWorker        │    │         DeployWorker            │    │
│   │                          │    │                                 │    │
│   │  poll release.build_     │    │  poll deployment.status         │    │
│   │  status = pending        │    │  = pending                      │    │
│   │                          │    │                                 │    │
│   │  1. GitHub API           │    │  1. vérifie release.build_      │    │
│   │     resolve git ref      │    │     status = succeeded          │    │
│   │  2. télécharge tarball   │    │  2. charge ProjectService +     │    │
│   │  3. docker build + push  │    │     env vars                    │    │
│   │  4. stocke image digest  │    │  3. envoie DeploymentSpec       │    │
│   └──────────┬───────────────┘    └─────────────┬───────────────────┘    │
└──────────────┼────────────────────────────────────┼─────────────────────┘
               │                                    │ HTTP POST /deployments
               │ docker build / push                ▼
               ▼                    ┌───────────────────────────────────────┐
┌─────────────────────┐             │          deploy-agent :8090           │
│   Docker Registry   │             │      (un par serveur Docker host)     │
│   (privé, local)    │             │                                       │
└─────────────────────┘             │  POST /deployments     → lance        │
                                    │  GET  /deployments/{id}→ poll santé   │
                                    │  DELETE /deployments/{id}→ stop       │
                                    └──────────────┬────────────────────────┘
                                                   │ docker pull / run / ps
                                                   ▼
                                    ┌──────────────────────────────────────┐
                                    │          Docker daemon               │
                                    │    (container applicatif en cours)   │
                                    └──────────────────────────────────────┘
```

## Pipeline Build

```
Opérateur                 control-plane-api         orchestrator-worker         GitHub / Registry
    │                            │                         │                          │
    │  POST /releases            │                         │                          │
    │  { version, gitRef }       │                         │                          │
    │ ─────────────────────────► │                         │                          │
    │                            │  GET /repos/.../commits │                          │
    │                            │ ───────────────────────────────────────────────►  │
    │                            │  ◄─────────────────── commitSHA ────────────────  │
    │                            │                         │                          │
    │                            │  INSERT release         │                          │
    │                            │  build_status=pending   │                          │
    │  ◄── 201 { id, status:pending } ──                  │                          │
    │                            │                         │                          │
    │  (poll GET /releases/{id}) │  ◄── poll ─────────────┤                          │
    │                            │                         │  GET /tarball/{sha}      │
    │                            │                         │ ────────────────────────►│
    │                            │                         │  ◄── archive.tar.gz ──── │
    │                            │                         │                          │
    │                            │                         │  docker build + push     │
    │                            │                         │ ────────────────────────►│
    │                            │                         │  ◄── image digest ─────  │
    │                            │                         │                          │
    │                            │  UPDATE release         │                          │
    │                            │  build_status=succeeded │                          │
    │  ◄── status: succeeded ────┤◄────────────────────── │                          │
```

## Pipeline Deploy

```
Opérateur           control-plane-api      orchestrator-worker       deploy-agent       Docker
    │                       │                       │                      │               │
    │  POST /deployments     │                       │                      │               │
    │  { releaseId, targetId}│                       │                      │               │
    │ ──────────────────────►│                       │                      │               │
    │                        │  INSERT deployment    │                      │               │
    │                        │  status=pending       │                      │               │
    │  ◄── 201 { id } ───────│                       │                      │               │
    │                        │                       │                      │               │
    │  (poll GET /deployments/{id})  ◄── poll ───────┤                      │               │
    │                        │                       │  vérifie             │               │
    │                        │                       │  buildStatus=succeeded               │
    │                        │                       │  charge service + env│               │
    │                        │                       │                      │               │
    │                        │                       │  POST /deployments   │               │
    │                        │                       │  { DeploymentSpec }  │               │
    │                        │                       │ ────────────────────►│               │
    │                        │                       │                      │  docker pull  │
    │                        │                       │                      │  docker run   │
    │                        │                       │                      │ ─────────────►│
    │                        │                       │  GET /deployments/id │               │
    │                        │                       │ ────────────────────►│               │
    │                        │                       │  ◄── status: healthy │               │
    │                        │                       │                      │               │
    │                        │  UPDATE status=healthy│                      │               │
    │  ◄── status: healthy ──│◄──────────────────────┤                      │               │
```

## Structure du projet

```text
dockyard/
  cmd/
    control-plane-api/         # API HTTP — état owner, publie les jobs
    orchestrator-worker/       # Build, release, deploy, rollback async
    deploy-agent/              # Agent HTTP sur chaque Docker host
  internal/
    domain/                    # Types métier purs (Project, Release, Deployment…)
    application/               # Services et cas d'usage
    ports/                     # Interfaces : repository, runtime, source, registry, agent, dns, routing
    adapters/
      postgres/                # CockroachDB — tous les repositories
      docker/                  # Runtime driver (docker CLI)
      dockerregistry/          # Build + push d'images (docker CLI)
      github/                  # Source provider (API GitHub v3)
      httpclient/              # Client HTTP vers les deploy-agents
      httpapi/                 # Router HTTP de l'API
  db/
    migrations/                # Migrations golang-migrate (CockroachDB)
  infra/
    local/                     # docker-compose : CockroachDB, Redis, Registry
  web/                         # Interface Next.js (à venir)
```

## Flux de déploiement

```
1. L'opérateur crée un projet (GitHub owner/repo, Dockerfile, branche)
2. Il déclenche un build → le worker résout le git ref via l'API GitHub,
   construit l'image Docker et la pousse dans le registry privé
3. Une Release est créée (image repository + tag + digest immuables)
4. L'opérateur crée un Deployment (release → runtime target)
5. Le worker détecte le deployment pending, contacte le deploy-agent
   de la cible via HTTP et lui envoie le DeploymentSpec
6. Le deploy-agent tire l'image et démarre le container Docker
7. Le worker poll le statut et met à jour le deployment (healthy / failed)
8. À terme : routage HTTP(S) via Nginx Proxy Manager + DNS via DuckDNS
```

## Commandes

```bash
# Infra locale (CockroachDB, Redis, Registry)
make local-infra-up
make local-infra-down
make local-infra-logs

# Migrations
make migrate-up
make migrate-down

# Services Go
make run-api       # Control Plane API  → :8080
make run-worker    # Orchestrator Worker
make run-agent     # Deploy Agent       → :8090  (requiert DOCKYARD_AGENT_KEY)

# Build, format, tests
make build
make fmt
make test
make test-integration   # requiert DOCKYARD_TEST_DSN ou l'infra locale up
```

## Variables d'environnement

| Variable | Service | Description |
|---|---|---|
| `DOCKYARD_DATABASE_URL` | api, worker | DSN CockroachDB (`postgresql://root@localhost:26257/dockyard?sslmode=disable`) |
| `DOCKYARD_API_ADDR` | api | Adresse d'écoute (défaut `:8080`) |
| `DOCKYARD_CORS_ALLOWED_ORIGINS` | api | Origines browser autorisées, séparées par des virgules (`http://localhost:3000,http://127.0.0.1:3000`) |
| `DOCKYARD_AGENT_ADDR` | agent | Adresse d'écoute (défaut `:8090`) |
| `DOCKYARD_AGENT_KEY` | agent, worker | Clé d'authentification inter-service (obligatoire sur l'agent) |
| `DOCKYARD_TEST_DSN` | tests | DSN pour les tests d'intégration Postgres |

## État du projet

### Implémenté

- [x] Domaine métier complet : `Project`, `Release`, `Deployment`, `RuntimeTarget`, `Domain`, `ProjectService`, `EnvironmentSet`, `EnvironmentVariable`
- [x] API HTTP CRUD pour toutes les ressources (projets, releases, déploiements, services, environnements, domaines)
- [x] Adapters Postgres (CockroachDB) pour tous les repositories
- [x] Migrations SQL (golang-migrate)
- [x] Tests d'intégration Postgres (race-free)
- [x] GitHub source adapter — résolution git ref + téléchargement d'archive tarball
- [x] Docker runtime driver — cycle de vie des containers via CLI
- [x] Docker registry builder — build + push d'images via CLI
- [x] Deploy-agent HTTP server — POST/GET/DELETE `/deployments/{id}`, auth, graceful shutdown
- [x] HTTP agent.Client — utilisé par l'orchestrator pour parler aux agents
- [x] BuildWorker — pipeline async build : poll pending releases → download → build → push → digest
- [x] DeployWorker — pipeline async deploy : poll pending deployments → DeploymentSpec complet (service + env vars) → dispatch agent → poll santé
- [x] `docs/FRONTEND.md` — spec complète de l'interface web

### Prochaines étapes

- [ ] Nginx Proxy Manager adapter (`routing.Provider`)
- [ ] DuckDNS adapter (`dns.Provider`)
- [ ] Interface web (voir `docs/FRONTEND.md`)
