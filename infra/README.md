# Dockyard Infra

Ce dossier regroupe l'infrastructure Dockyard, son layout, et la procedure de deploiement recommandee pour une V1 simple sur trois serveurs Docker.

## Vue d'ensemble

- `foundation/` : briques partagees qui ne sont pas gerees par Dockyard lui-meme
- `platform/` : services Dockyard (`control-plane-api`, `orchestrator-worker`, `web`)
- `agents/` : agent de deploiement a installer sur chaque host cible
- `local/` : stack de support pour le developpement local

```text
infra/
  foundation/
    cockroach/
    redis/
    registry/
    traefik/
  platform/
    dockyard/
  agents/
    deploy-agent/
  local/
    docker-compose.yml
```

## Regles de deploiement

- `foundation/` reste hors du perimetre de gestion de Dockyard au demarrage
- `platform/` heberge Dockyard lui-meme
- `agents/` est deploye sur tous les serveurs qui doivent recevoir des applications
- `local/` sert uniquement au developpement local
- CockroachDB ne doit pas etre expose derriere Traefik

## Repartition recommandee

### `server-1`

- `cockroach-1`
- `redis`
- `registry`
- `traefik`
- `dockyard-control-plane-api`
- `dockyard-web`
- `deploy-agent`

### `server-2`

- `cockroach-2`
- `dockyard-orchestrator-worker`
- `deploy-agent`

### `server-3`

- `cockroach-3`
- `deploy-agent`

Cette repartition garde la V1 simple :

- la base reste distribuee sur trois hosts
- Redis, Registry et Traefik restent en instance simple
- l'API et le web sont regroupes sur un host
- le worker est separe

## Reseaux et stockage

Reseaux Docker a creer selon les hosts concernes :

- `dockyard_foundation`
- `dockyard_edge`
- `dockyard_platform`

Volumes persistants recommandes :

- CockroachDB : `/opt/dockyard/cockroach`
- Redis : `/opt/dockyard/redis`
- Registry : `/opt/dockyard/registry`
- Traefik ACME : `/opt/dockyard/traefik/acme`

## Ordre de bootstrap

1. demarrer les trois noeuds CockroachDB
2. initialiser le cluster
3. creer la base `dockyard`
4. deployer Redis
5. deployer Registry
6. deployer Traefik
7. deployer l'API, le worker et le web
8. deployer l'agent sur chaque host

## Deploiement

### `server-1`

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

### `server-2`

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

### `server-3`

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

Une seule fois, apres le demarrage des trois noeuds, depuis `server-1` :

```bash
docker exec -it cockroach-1 cockroach init
docker exec -it cockroach-1 cockroach sql -e "CREATE DATABASE dockyard;"
```

## Mises a jour

Mettre a jour web + API sur `server-1` :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build control-plane-api web
```

Mettre a jour le worker sur `server-2` :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build orchestrator-worker
```

Mettre a jour un agent sur le host cible :

```bash
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

## Developpement local

Lancer les services de support :

```bash
make local-infra-up
```

Lancer le backend Go en local :

```bash
make run-api
make run-worker
make run-agent
```

Lancer le frontend :

```bash
make web-dev
```
