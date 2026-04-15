# Deploy Agent

Un agent doit tourner sur chaque host Docker cible.

## Prerequis

```bash
docker network create dockyard_platform
```

## Lancement

```bash
cd infra/agents/deploy-agent
docker compose -f compose.yml up -d --build
```

## Securite

Ce conteneur monte `/var/run/docker.sock`.
Il doit rester sur un reseau prive et ne pas etre expose publiquement.
