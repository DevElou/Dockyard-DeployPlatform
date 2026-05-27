# Dockyard Infra

Ce dossier regroupe l'infrastructure Dockyard, son layout, et la procedure de deploiement recommandee.

## Vue d'ensemble

- `foundation/` : briques partagees qui ne sont pas gerees par Dockyard lui-meme
- `platform/` : services Dockyard (`control-plane-api`, `orchestrator-worker`)
- `agents/` : agent de deploiement a installer sur chaque host cible
- `local/` : stack de support pour le developpement local

```text
infra/
  foundation/
    cockroach/
    redis/
    registry/
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
- CockroachDB ne doit pas etre expose derriere Nginx Proxy Manager
- Nginx Proxy Manager est une infrastructure existante, non deployee par Dockyard

## Repartition

### `server-1` — Control plane complet

- `cockroach` (noeud unique, `start-single-node`)
- `redis`
- `registry` (Docker registry prive)
- `dockyard-control-plane-api` :8080
- `dockyard-orchestrator-worker`
- `deploy-agent` :8090

### `server-2` et `server-3` — Agents uniquement

- `deploy-agent` :8090

## Reseaux et stockage

Reseaux Docker a creer sur `server-1` :

- `dockyard_foundation`
- `dockyard_edge`
- `dockyard_platform`

Reseaux Docker a creer sur `server-2` et `server-3` :

- `dockyard_platform`

Volumes persistants :

- CockroachDB : `/opt/dockyard/cockroach`
- Redis : `/opt/dockyard/redis`
- Registry : `/opt/dockyard/registry`

## Ordre de bootstrap

1. demarrer CockroachDB en mode single-node (pas d'init cluster requis)
2. creer la base `dockyard` et appliquer les migrations
3. deployer Redis
4. deployer Registry
5. deployer l'API et le worker
6. deployer l'agent sur chaque host cible

## Deploiement

### `server-1`

Creation des reseaux :

```bash
docker network create dockyard_foundation || true
docker network create dockyard_edge || true
docker network create dockyard_platform || true
```

CockroachDB (single-node, pas de cluster) :

```bash
cd infra/foundation/cockroach
docker compose --env-file ../../../.env -f single.compose.yml up -d
docker exec -it cockroach cockroach sql --insecure \
  -e "CREATE DATABASE IF NOT EXISTS dockyard;"
```

Redis :

```bash
cd infra/foundation/redis
docker compose --env-file ../../../.env up -d
```

Registry :

```bash
cd infra/foundation/registry
docker compose --env-file ../../../.env up -d
```

Migrations :

```bash
make migrate-up
```

Dockyard API + worker :

```bash
make deploy-platform
```

Deploy agent :

```bash
make deploy-agent
```

### `server-2` et `server-3` — Agent uniquement

Creation des reseaux :

```bash
docker network create dockyard_platform || true
```

Deploy agent :

```bash
make deploy-agent
```

## Mises a jour

Mettre a jour API + worker sur `server-1` :

```bash
make deploy-platform
```

Mettre a jour un agent :

```bash
make deploy-agent
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
