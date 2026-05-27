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
make deploy-platform
```

## Lancement partiel

Frontend seul :

```bash
make deploy-web
```

Etat des services :

```bash
make deploy-ps
```
