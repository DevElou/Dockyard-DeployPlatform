# Dockyard Infra

Cette arborescence separe clairement :

- `foundation` : briques d'infrastructure partagees
- `platform` : services Dockyard
- `agents` : agent de deploiement installe sur chaque host
- `local` : stack de support pour le developpement local

## Structure

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

## Regle de deploiement

- `foundation/` n'est pas gere par Dockyard lui-meme
- `platform/` heberge Dockyard
- `agents/` est deploye sur chaque serveur cible
- `local/` sert uniquement au developpement local

## Ordre de demarrage

1. `foundation/cockroach`
2. `foundation/redis`
3. `foundation/registry`
4. `foundation/traefik`
5. `platform/dockyard`
6. `agents/deploy-agent`

## Commandes de deploiement

### 1. CockroachDB

Sur `server-1` :

```bash
cd infra/foundation/cockroach
docker compose -f server-1.compose.yml up -d
```

Sur `server-2` :

```bash
cd infra/foundation/cockroach
docker compose -f server-2.compose.yml up -d
```

Sur `server-3` :

```bash
cd infra/foundation/cockroach
docker compose -f server-3.compose.yml up -d
```

Puis, une seule fois depuis `server-1` :

```bash
docker exec -it cockroach-1 cockroach init
docker exec -it cockroach-1 cockroach sql -e "CREATE DATABASE dockyard;"
```

### 2. Redis

Sur le host qui portera Redis :

```bash
cd infra/foundation/redis
docker compose -f compose.yml up -d
```

### 3. Docker Registry

Sur le host qui portera le registry :

```bash
cd infra/foundation/registry
docker compose -f compose.yml up -d
```

### 4. Traefik

Sur le host qui portera Traefik :

```bash
docker network create dockyard_edge
cd infra/foundation/traefik
docker compose -f compose.yml up -d
```

### 5. Dockyard Platform

Sur le host qui portera les services Dockyard :

```bash
docker network create dockyard_platform
docker network create dockyard_edge
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build
```

### 6. Deploy Agent

Sur chaque host cible :

```bash
docker network create dockyard_platform
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

## Developpement local

Pour lancer les services de support :

```bash
make local-infra-up
```

Pour lancer le backend Go en local :

```bash
make run-api
make run-worker
make run-agent
```

Pour lancer le frontend :

```bash
make web-dev
```
