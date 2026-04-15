# Dockyard

Plateforme privee de deploiement type "mini Vercel self-hosted", pensee pour une V1 simple sur Docker avec une trajectoire claire vers Kubernetes.

## Documents de base

- [Architecture cible](./docs/architecture.md)
- [ADR 0001 - Architecture V1](./docs/adrs/0001-architecture-v1.md)
- [Guide architecture et Go](./docs/go-architecture-guide.md)
- [Schéma SQL initial CockroachDB](./db/schema.sql)
- [Contrats Go initiaux](./internal/ports/README.md)

## Stack retenue

- Frontend : Next.js
- Backend API : Go
- Worker d'orchestration : Go
- Agent de deploiement : Go
- Base principale : CockroachDB
- Queue : Redis
- Edge : Traefik
- Registry : Docker Registry prive

## Structure cible

```text
dockyard/
  web/
  cmd/
    control-plane-api/
    orchestrator-worker/
    deploy-agent/
  internal/
    domain/
    application/
    ports/
    adapters/
  db/
  docs/
```

## Commandes utiles

```bash
make run-api
make run-worker
make run-agent
make build
make test
```

Pour le frontend :

```bash
cd web
npm install
npm run dev
```
