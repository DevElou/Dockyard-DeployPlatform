# Dockyard Deployment

Ce document donne une proposition simple pour deployer Dockyard sur tes trois serveurs Docker.

## Repartition recommandee

### Serveur 1

- `cockroach-1`
- `redis`
- `registry`
- `traefik`
- `dockyard-control-plane-api`
- `dockyard-web`
- `deploy-agent`

### Serveur 2

- `cockroach-2`
- `dockyard-orchestrator-worker`
- `deploy-agent`

### Serveur 3

- `cockroach-3`
- `deploy-agent`

Cette repartition reste simple :
- la base est distribuee sur trois hosts
- Redis, Traefik et Registry restent en instance simple en V1
- l'API et le web sont regroupes sur un host
- le worker est separe

## Reseaux Docker

Conserve des reseaux explicites :

- `dockyard_foundation`
- `dockyard_edge`
- `dockyard_platform`

En V1, le plus simple est de creer les reseaux externes a la main sur les hosts qui en ont besoin :

```bash
docker network create dockyard_foundation
docker network create dockyard_edge
docker network create dockyard_platform
```

## Stockage persistant

Volumes recommandés sur chaque host :

- CockroachDB : `/opt/dockyard/cockroach`
- Redis : `/opt/dockyard/redis`
- Registry : `/opt/dockyard/registry`
- Traefik ACME : `/opt/dockyard/traefik/acme`

## Flux de bootstrap

1. deployer les trois noeuds CockroachDB
2. initialiser le cluster
3. creer la base `dockyard`
4. deployer Redis
5. deployer Registry
6. deployer Traefik
7. builder et deployer l'API, le worker et le web
8. deployer l'agent sur chaque host

## Commandes exactes

### Server 1

Creation des reseaux :

```bash
docker network create dockyard_foundation || true
docker network create dockyard_edge || true
docker network create dockyard_platform || true
```

CockroachDB :

```bash
cd infra/foundation/cockroach
docker compose -f server-1.compose.yml up -d
```

Redis :

```bash
cd infra/foundation/redis
docker compose -f compose.yml up -d
```

Registry :

```bash
cd infra/foundation/registry
docker compose -f compose.yml up -d
```

Traefik :

```bash
cd infra/foundation/traefik
docker compose -f compose.yml up -d
```

Dockyard web + API :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build control-plane-api web
```

Deploy agent :

```bash
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

### Server 2

Creation des reseaux :

```bash
docker network create dockyard_foundation || true
docker network create dockyard_platform || true
```

CockroachDB :

```bash
cd infra/foundation/cockroach
docker compose -f server-2.compose.yml up -d
```

Dockyard worker :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build orchestrator-worker
```

Deploy agent :

```bash
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

### Server 3

Creation des reseaux :

```bash
docker network create dockyard_foundation || true
docker network create dockyard_platform || true
```

CockroachDB :

```bash
cd infra/foundation/cockroach
docker compose -f server-3.compose.yml up -d
```

Deploy agent :

```bash
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

### Initialisation CockroachDB

Une seule fois, apres demarrage des trois noeuds, depuis `server-1` :

```bash
docker exec -it cockroach-1 cockroach init
docker exec -it cockroach-1 cockroach sql -e "CREATE DATABASE dockyard;"
```

## Commandes de mise a jour

### Mettre a jour web + API

Sur `server-1` :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build control-plane-api web
```

### Mettre a jour le worker

Sur `server-2` :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build orchestrator-worker
```

### Mettre a jour un agent

Sur le host cible :

```bash
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

## Important

- ne fais pas gerer `foundation/` par Dockyard au debut
- garde l'agent sur tous les hosts qui doivent recevoir des apps
- ne mets pas CockroachDB derriere Traefik
