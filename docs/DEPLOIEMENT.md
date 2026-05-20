# Déploiement Dockyard

Ce guide couvre le déploiement de Dockyard sur trois serveurs Docker dans un homelab ESXi.

> **Nginx Proxy Manager** est une infrastructure existante — Dockyard ne le déploie pas.
> Il sera utilisé automatiquement par Dockyard pour le routage HTTP(S) une fois
> l'adapter `routing.Provider` implémenté (P2).

## Prérequis

| Outil | Version | Usage |
|---|---|---|
| Docker + Docker Compose | 24+ | Tous les services |
| `golang-migrate` | v4 | Migrations SQL |
| `openssl` | — | Génération de la clé agent |

Installer `golang-migrate` :
```bash
go install -tags 'cockroachdb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## Répartition des services

```
server-1  ─── CockroachDB (nœud 1)
           ── Redis
           ── Registry Docker privé
           ── control-plane-api   :8080
           ── deploy-agent        :8090

server-2  ─── CockroachDB (nœud 2)
           ── orchestrator-worker
           ── deploy-agent        :8090

server-3  ─── CockroachDB (nœud 3)
           ── deploy-agent        :8090

(existant) ── Nginx Proxy Manager  — géré séparément, non déployé par Dockyard
```

---

## Configuration

### 1. Créer le fichier `.env`

Copier `.env.example` à la racine du dépôt :

```bash
cp .env.example .env
```

Remplir les valeurs obligatoires :

| Variable | Description |
|---|---|
| `DOCKYARD_DATABASE_URL` | DSN CockroachDB (pgx) |
| `DOCKYARD_GITHUB_TOKEN` | Token GitHub scope `repo` — [générer ici](https://github.com/settings/tokens) |
| `DOCKYARD_REGISTRY_URL` | URL du registry sans schéma (ex: `server-1.local:5000`) |
| `DOCKYARD_AGENT_KEY` | Clé partagée API ↔ agents — générer avec `openssl rand -hex 32` |

Générer la clé agent :
```bash
openssl rand -hex 32
# coller le résultat dans DOCKYARD_AGENT_KEY
```

---

## Bootstrap — une seule fois

### Étape 1 — Réseaux Docker

À créer sur **chaque serveur** avant de lancer quoi que ce soit :

```bash
# server-1
docker network create dockyard_foundation || true
docker network create dockyard_platform   || true
docker network create dockyard_edge       || true

# server-2 et server-3
docker network create dockyard_foundation || true
docker network create dockyard_platform   || true
```

> `dockyard_edge` est le réseau partagé entre l'API et Nginx Proxy Manager.
> Si NPM tourne déjà sur un réseau existant, adapter le nom dans
> `infra/platform/dockyard/compose.yml`.

### Étape 2 — CockroachDB

Démarrer un nœud par serveur. Sur **chaque serveur**, exporter l'adresse annoncée :

```bash
# server-1
export COCKROACH_ADVERTISE_ADDR=server-1.local
cd infra/foundation/cockroach
docker compose --env-file ../../../.env -f server-1.compose.yml up -d

# server-2
export COCKROACH_ADVERTISE_ADDR=server-2.local
docker compose --env-file ../../../.env -f server-2.compose.yml up -d

# server-3
export COCKROACH_ADVERTISE_ADDR=server-3.local
docker compose --env-file ../../../.env -f server-3.compose.yml up -d
```

Initialiser le cluster — **une seule fois depuis server-1** :

```bash
docker exec -it cockroach-1 cockroach init --insecure
docker exec -it cockroach-1 cockroach sql --insecure \
  -e "CREATE DATABASE IF NOT EXISTS dockyard;"
```

Vérifier que les trois nœuds sont actifs :
```bash
docker exec -it cockroach-1 cockroach node status --insecure
# doit afficher 3 nœuds avec is_live=true
```

### Étape 3 — Redis et Registry (server-1)

```bash
cd infra/foundation/redis
docker compose --env-file ../../../.env up -d

cd ../registry
docker compose --env-file ../../../.env up -d
```

Autoriser le registry insécure sur **chaque host Docker** qui buildera ou tirera des images :
```bash
# /etc/docker/daemon.json
{ "insecure-registries": ["server-1.local:5000"] }
sudo systemctl restart docker
```

### Étape 4 — Migrations SQL

Depuis la machine de déploiement (accès réseau vers server-1:26257) :

```bash
make migrate-up
# ou directement :
migrate -path db/migrations \
        -database "cockroachdb://root@server-1.local:26257/dockyard?sslmode=disable" \
        up
```

### Étape 5 — Services Dockyard

**server-1** — API :
```bash
cd infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build control-plane-api
```

**server-2** — Worker :
```bash
cd infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build orchestrator-worker
```

**Chaque serveur** — Agent :
```bash
cd infra/agents/deploy-agent
docker compose --env-file ../../../.env up -d --build
```

---

## Vérification

```bash
# L'API répond
curl http://server-1.local:8080/healthz
# → {"status":"ok"}

# Les logs du worker
docker logs dockyard-orchestrator-worker -f

# Les logs de l'agent
docker logs dockyard-deploy-agent -f
```

---

## Mises à jour

```bash
# API (server-1)
cd infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build control-plane-api

# Worker (server-2)
docker compose --env-file ../../../.env up -d --build orchestrator-worker

# Agent sur un host cible
cd infra/agents/deploy-agent
docker compose --env-file ../../../.env up -d --build
```

Appliquer de nouvelles migrations avant de redémarrer les services :
```bash
make migrate-up
```

---

## Développement local

```bash
# 1. Démarrer CockroachDB + Redis + Registry en local
make local-infra-up

# 2. Créer la base et appliquer les migrations
make db-create
make migrate-up

# 3. Exporter les variables requises
export DOCKYARD_GITHUB_TOKEN=ghp_xxx
export DOCKYARD_REGISTRY_URL=localhost:5000
export DOCKYARD_AGENT_KEY=dev-key-local

# 4. Lancer les services dans trois terminaux
make run-api      # terminal 1 → :8080
make run-worker   # terminal 2
make run-agent    # terminal 3 → :8090
```

Tests d'intégration (nécessite l'infra locale) :
```bash
make test-integration
```

---

## Troubleshooting

**Le worker plante au démarrage avec "DOCKYARD_GITHUB_TOKEN is required"**
→ La variable n'est pas définie ou vide dans `.env`. Vérifier avec `docker compose config`.

**`migrate up` échoue avec "no such host"**
→ Le nom `server-1.local` n'est pas résolu. Utiliser l'IP directement dans `MIGRATE_URL`.

**Le registry retourne une erreur lors du push**
→ Vérifier que `insecure-registries` est configuré dans `/etc/docker/daemon.json` sur tous les hosts.

**CockroachDB : un nœud ne rejoint pas le cluster**
→ Vérifier que `COCKROACH_JOIN` contient les bonnes adresses et que le port 26257 est ouvert entre les serveurs.

**L'API ne répond pas derrière NPM**
→ Dans NPM, créer un proxy host pointant vers `server-1.local:8080`.
   S'assurer que le container `dockyard-control-plane-api` est sur le réseau `dockyard_edge`
   et que NPM est sur ce même réseau.
