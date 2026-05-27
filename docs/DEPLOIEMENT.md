# Déploiement Dockyard — Guide complet

Ce guide couvre le déploiement de Dockyard sur un homelab ESXi avec trois serveurs Docker,
ainsi que le développement local via Docker Compose.

---

## Table des matières

1. [Architecture](#1-architecture)
2. [Prérequis](#2-prérequis)
3. [Configuration](#3-configuration)
4. [Bootstrap server-1 (première fois)](#4-bootstrap-server-1-première-fois)
5. [Bootstrap server-2 et server-3](#5-bootstrap-server-2-et-server-3)
6. [Accès à l'interface web](#6-accès-à-linterface-web)
7. [Enregistrer les runtime targets](#7-enregistrer-les-runtime-targets)
8. [Mises à jour](#8-mises-à-jour)
9. [Développement local](#9-développement-local)
10. [Troubleshooting](#10-troubleshooting)

---

## 1. Architecture

```
┌─────────────────────────────────────────────────────────┐
│  server-1 (control plane)                               │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ CockroachDB  │  │    Redis     │  │   Registry   │  │
│  │   :26257     │  │    :6379     │  │    :5000     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                         │
│  ┌──────────────────────┐  ┌──────────────────────────┐ │
│  │  control-plane-api   │  │   orchestrator-worker    │ │
│  │       :8080          │  │  (build + deploy async)  │ │
│  └──────────────────────┘  └──────────────────────────┘ │
│                                                         │
│  ┌──────────────┐  ┌──────────────────────────────────┐ │
│  │     web      │  │          deploy-agent            │ │
│  │    :3000     │  │              :8090               │ │
│  └──────────────┘  └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘

┌────────────────────┐     ┌────────────────────┐
│      server-2      │     │      server-3      │
│  ┌──────────────┐  │     │  ┌──────────────┐  │
│  │ deploy-agent │  │     │  │ deploy-agent │  │
│  │    :8090     │  │     │  │    :8090     │  │
│  └──────────────┘  │     │  └──────────────┘  │
└────────────────────┘     └────────────────────┘

(existant) Nginx Proxy Manager — géré séparément, non déployé par Dockyard
```

### Réseaux Docker

| Réseau | Membres |
|---|---|
| `dockyard_foundation` | CockroachDB |
| `dockyard_platform` | API, Worker, Web, Agents |
| `dockyard_edge` | API, Web, Nginx Proxy Manager |

---

## 2. Prérequis

### Tous les serveurs

- Docker Engine 24+ avec Docker Compose V2 (`docker compose version`)
- Git

### server-1 uniquement

- Accès réseau depuis la machine de déploiement vers `:26257` (migrations)
- Socket Docker accessible depuis les containers (`/var/run/docker.sock`)

### Machine de déploiement (CI ou poste dev)

`golang-migrate` v4 avec driver CockroachDB :

```bash
go install -tags 'cockroachdb' \
  github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

> Alternative sans Go : utiliser le container `migrate/migrate` (cf. section 4.6).

---

## 3. Configuration

### 3.1 Créer le fichier `.env`

```bash
cp .env.example .env
```

### 3.2 Variables obligatoires

| Variable | Description | Exemple |
|---|---|---|
| `DOCKYARD_DATABASE_URL` | DSN pgx vers CockroachDB | `postgresql://root@server-1.local:26257/dockyard?sslmode=disable` |
| `MIGRATE_URL` | DSN golang-migrate | `cockroachdb://root@server-1.local:26257/dockyard?sslmode=disable` |
| `DOCKYARD_GITHUB_TOKEN` | Token GitHub scope `repo` | `ghp_xxxx` |
| `DOCKYARD_REGISTRY_URL` | Registry sans schéma | `server-1.local:5000` |
| `DOCKYARD_AGENT_KEY` | Clé partagée API ↔ agents | *(générer ci-dessous)* |
| `NEXT_PUBLIC_API_BASE_URL` | URL de l'API vue du navigateur | `http://server-1.local:8080` |

**Générer la clé agent** (une seule fois, même valeur sur tous les serveurs) :

```bash
openssl rand -hex 32
# → coller dans DOCKYARD_AGENT_KEY
```

**Générer le token GitHub** → [github.com/settings/tokens](https://github.com/settings/tokens)  
Scope requis : `repo` (lecture seule sur les dépôts à builder).

### 3.3 Variables NPM (optionnel)

Laisser vide pour désactiver le routing automatique :

```dotenv
DOCKYARD_NPM_URL=http://server-1.local:81
DOCKYARD_NPM_IDENTITY=admin@example.com
DOCKYARD_NPM_SECRET=your-npm-password
DOCKYARD_NPM_DEFAULT_FORWARD_SCHEME=http
```

---

## 4. Bootstrap server-1 (première fois)

> Ces étapes sont **à exécuter une seule fois** lors de l'installation initiale.
> Pour les mises à jour, voir la [section 8](#8-mises-à-jour).

### 4.1 Cloner le dépôt

```bash
git clone https://github.com/elouan/dockyard.git /opt/dockyard/repo
cd /opt/dockyard/repo
cp .env.example .env
# → éditer .env avec les vraies valeurs
```

### 4.2 Créer les réseaux Docker

```bash
docker network create dockyard_foundation || true
docker network create dockyard_platform   || true
docker network create dockyard_edge       || true
```

### 4.3 Démarrer CockroachDB

```bash
cd infra/foundation/cockroach
docker compose --env-file ../../../.env -f single.compose.yml up -d
```

Attendre ~10 s que CockroachDB soit prêt, puis créer la base :

```bash
docker exec -it cockroach cockroach sql --insecure \
  -e "CREATE DATABASE IF NOT EXISTS dockyard;"
```

### 4.4 Démarrer Redis et le Registry

```bash
cd /opt/dockyard/repo/infra/foundation/redis
docker compose --env-file ../../../.env up -d

cd /opt/dockyard/repo/infra/foundation/registry
docker compose --env-file ../../../.env up -d
```

### 4.5 Autoriser le registry insécurisé

Sur **chaque serveur** (server-1, server-2, server-3) :

```bash
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "insecure-registries": ["server-1.local:5000"]
}
EOF
sudo systemctl restart docker
```

> Remplacer `server-1.local` par l'IP ou le hostname réel de votre server-1.

### 4.6 Appliquer les migrations SQL

**Option A — avec `golang-migrate` installé localement :**

```bash
cd /opt/dockyard/repo
make migrate-up
```

**Option B — sans rien installer (via Docker) :**

```bash
docker run --rm \
  --network host \
  -v "/opt/dockyard/repo/db/migrations:/migrations" \
  migrate/migrate:latest \
  -path /migrations \
  -database "cockroachdb://root@localhost:26257/dockyard?sslmode=disable" \
  up
```

### 4.7 Déployer l'API, le Worker et le Frontend

```bash
cd /opt/dockyard/repo/infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build
```

Trois containers démarrent :
- `dockyard-control-plane-api` → `:8080`
- `dockyard-orchestrator-worker`
- `dockyard-web` → `:3000`

> **Important :** Le frontend est compilé avec `NEXT_PUBLIC_API_BASE_URL` baked dans le bundle.
> Si cette variable change (ex. passage à HTTPS), reconstruire avec `--build web`.

### 4.8 Déployer l'agent sur server-1

```bash
cd /opt/dockyard/repo/infra/agents/deploy-agent
docker compose --env-file ../../../.env up -d --build
```

### 4.9 Vérifier l'installation

```bash
# API
curl http://server-1.local:8080/healthz
# → {"status":"ok"}

# Interface web
curl -I http://server-1.local:3000
# → HTTP/1.1 200 OK

# Agent
curl http://server-1.local:8090/healthz
# → {"status":"ok"}

# Logs
docker logs dockyard-control-plane-api   --tail=20
docker logs dockyard-orchestrator-worker --tail=20
docker logs dockyard-deploy-agent        --tail=20
```

---

## 5. Bootstrap server-2 et server-3

Sur chaque serveur secondaire (agent seulement) :

```bash
# 1. Cloner le dépôt
git clone https://github.com/elouan/dockyard.git /opt/dockyard/repo
cd /opt/dockyard/repo

# 2. Fichier .env minimal (seul DOCKYARD_AGENT_KEY est requis)
cp .env.example .env
# → renseigner DOCKYARD_AGENT_KEY avec la MÊME valeur que sur server-1

# 3. Réseau
docker network create dockyard_platform || true

# 4. Registry insécurisé
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "insecure-registries": ["server-1.local:5000"]
}
EOF
sudo systemctl restart docker

# 5. Agent
cd infra/agents/deploy-agent
docker compose --env-file ../../../.env up -d --build

# 6. Vérifier
curl http://localhost:8090/healthz
# → {"status":"ok"}
```

---

## 6. Accès à l'interface web

### Accès direct

Ouvrir **http://server-1.local:3000** dans le navigateur.

### Via Nginx Proxy Manager (recommandé)

Dans l'interface NPM (`http://server-1.local:81`) :

**Exposer le frontend :**

| Champ | Valeur |
|---|---|
| Domain names | `dockyard.home` |
| Forward Hostname/IP | `dockyard-web` |
| Forward Port | `3000` |

> `dockyard-web` est le nom du container sur le réseau `dockyard_edge` partagé avec NPM.

**Exposer l'API (si accès externe souhaité) :**

| Champ | Valeur |
|---|---|
| Domain names | `api.dockyard.home` |
| Forward Hostname/IP | `dockyard-control-plane-api` |
| Forward Port | `8080` |

> Si l'API est exposée via un domaine HTTPS, mettre à jour `NEXT_PUBLIC_API_BASE_URL=https://api.dockyard.home`
> dans `.env` et reconstruire l'image web :
> ```bash
> docker compose --env-file ../../../.env up -d --build web
> ```

---

## 7. Enregistrer les runtime targets

Une fois l'interface accessible, enregistrer chaque serveur comme cible de déploiement.

1. Aller dans **Settings → Runtime Targets**
2. Cliquer **Add target**
3. Renseigner :
   - **Name** : `server-1`, `server-2`, `server-3`
   - **Endpoint** : URL de l'agent (ex. `http://server-1.local:8090`)
   - **Agent key** : valeur de `DOCKYARD_AGENT_KEY`
4. Activer le toggle

> L'endpoint doit être **accessible depuis l'intérieur du container `orchestrator-worker`**.
> Utiliser les hostnames réseau (pas `localhost`) pour les agents distants.
>
> Tester depuis le worker :
> ```bash
> docker exec dockyard-orchestrator-worker \
>   wget -qO- http://server-2.local:8090/healthz
> ```

---

## 8. Mises à jour

### 8.1 API, Worker et Frontend (server-1)

```bash
cd /opt/dockyard/repo
git pull

# Migrations si de nouvelles existent
make migrate-up

# Reconstruire et redémarrer
cd infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build
```

### 8.2 Agent (chaque serveur cible)

```bash
cd /opt/dockyard/repo
git pull

cd infra/agents/deploy-agent
docker compose --env-file ../../../.env up -d --build
```

### 8.3 Rollback rapide

```bash
# Revenir à un commit précédent
git checkout <commit-ou-tag>

# Reconstruire
cd infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build
```

---

## 9. Développement local

### Option A — Stack Docker complète (recommandée)

Tout en une commande, sans rien installer :

```bash
cp .env.dev.example .env.dev
# → renseigner DOCKYARD_GITHUB_TOKEN dans .env.dev

# Docker Desktop : Settings → Docker Engine → ajouter :
# "insecure-registries": ["localhost:5000"]

make dev-build   # premier lancement (~5 min)
```

| Commande | Action |
|---|---|
| `make dev-up` | Démarrer sans rebuild |
| `make dev-build` | Démarrer + rebuild les images |
| `make dev-down` | Arrêter (volumes conservés) |
| `make dev-reset` | Arrêter + supprimer tous les volumes |
| `make dev-logs` | Suivre les logs en direct |
| `make dev-ps` | État des containers |

Ports locaux :

| Service | URL |
|---|---|
| Interface web | http://localhost:3000 |
| API | http://localhost:8080 |
| UI CockroachDB | http://localhost:8081 |
| Deploy Agent | http://localhost:8090 |
| Registry | localhost:5000 |

### Option B — Services Go natifs (hot reload)

```bash
make local-infra-up    # CockroachDB + Redis + Registry
make db-create
make migrate-up

export $(grep -v '^#' .env | xargs)

make run-api      # terminal 1 → :8080
make run-worker   # terminal 2
make run-agent    # terminal 3 → :8090
make web-dev      # terminal 4 → :3000
```

### Tests

```bash
make test               # tests unitaires (sans DB)
make test-integration   # tests d'intégration (nécessite make local-infra-up)
```

---

## 10. Troubleshooting

### `DOCKYARD_GITHUB_TOKEN is required` au démarrage

La variable est absente ou vide dans `.env` :

```bash
docker compose --env-file .env config | grep GITHUB
```

---

### `migrate up` — `no such host`

Le hostname ne se résout pas. Utiliser l'IP directement dans `MIGRATE_URL` :

```dotenv
MIGRATE_URL=cockroachdb://root@192.168.1.10:26257/dockyard?sslmode=disable
```

---

### Push vers le registry — `http: server gave HTTP response to HTTPS client`

Le registry n'est pas dans les registries insécurisés du daemon Docker :

```bash
cat /etc/docker/daemon.json
# Doit contenir : "insecure-registries": ["server-1.local:5000"]
sudo systemctl restart docker
```

---

### L'API ne répond pas derrière NPM

Vérifier que le container est bien sur le réseau `dockyard_edge` :

```bash
docker inspect dockyard-control-plane-api \
  --format '{{json .NetworkSettings.Networks}}' | jq 'keys'
# doit contenir "dockyard_edge"
```

Si absent :

```bash
docker network connect dockyard_edge dockyard-control-plane-api
```

---

### Le deploy-agent n'est pas joignable depuis le worker

Tester la connectivité depuis l'intérieur du container worker :

```bash
docker exec dockyard-orchestrator-worker \
  wget -qO- http://server-2.local:8090/healthz
```

Si ça échoue : vérifier le firewall, les hostnames, et que les deux containers sont sur `dockyard_platform`.

---

### Reconstruire le frontend après changement de `NEXT_PUBLIC_API_BASE_URL`

Cette variable est baked dans le bundle JavaScript au build. Un `docker compose up -d` sans `--build` **ne suffit pas** :

```bash
cd infra/platform/dockyard
docker compose --env-file ../../../.env up -d --build web
```
