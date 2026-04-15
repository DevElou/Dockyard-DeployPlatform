# Dockyard Platform

Ce compose deploie les services principaux :

- `control-plane-api`
- `orchestrator-worker`
- `web`

## Prerequis

```bash
docker network create dockyard_platform
docker network create dockyard_edge
```

## Lancement

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build
```

## Lancement partiel

API + web :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build control-plane-api web
```

Worker seul :

```bash
cd infra/platform/dockyard
docker compose -f compose.yml up -d --build orchestrator-worker
```
