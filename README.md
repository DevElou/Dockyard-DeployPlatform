# Dockyard

Dockyard est une plateforme privee de deploiement inspiree de Vercel, concue pour piloter une infrastructure personnelle ou homelab. L'objectif est de fournir une interface web unique pour connecter un projet, construire son image Docker, la deployer sur des serveurs Docker, puis gerer son exposition HTTP(S).

Le cas d'usage cible est une infra ESXi avec plusieurs VM ou serveurs faisant tourner Docker. Dockyard agit comme un control plane : il garde l'etat des projets, orchestre les builds et deploiements, puis delegue l'execution a des agents installes sur les hosts Docker.

## Objectifs

- Enregistrer un projet depuis un repository contenant une application Docker.
- Construire une image versionnee a partir du repo et la pousser dans un registry prive.
- Deployer automatiquement cette image sur un ou plusieurs serveurs Docker.
- Suivre les releases, deployments, statuts et historiques depuis une interface web.
- Permettre le rollback vers une release precedente.
- Piloter l'exposition des applications depuis l'interface, idealement via Nginx Proxy Manager local.

## Architecture V1

```text
Next.js Web UI
    |
Control Plane API (Go)
    |
CockroachDB + Redis
    |
Orchestrator Worker (Go)
    |
Deploy Agents (Go) -> Docker hosts on ESXi
    |
Nginx Proxy Manager / edge routing
```

La logique metier reste separee des outils concrets. Docker, GitHub, le registry et le reverse proxy doivent rester derriere des interfaces pour pouvoir changer d'implementation plus tard.

## Stack

- Frontend : Next.js
- Backend API : Go
- Worker d'orchestration : Go
- Agent de deploiement : Go
- Base principale : CockroachDB
- Queue : Redis
- Runtime cible : Docker
- Registry : Docker Registry prive
- Reverse proxy cible : Nginx Proxy Manager, via integration API si possible

## Structure du projet

```text
dockyard/
  web/                         # interface Next.js
  cmd/
    control-plane-api/         # API Go
    orchestrator-worker/       # worker de build/deploiement
    deploy-agent/              # agent installe sur les hosts Docker
  internal/
    domain/                    # modele metier
    application/               # services et cas d'usage
    ports/                     # interfaces vers l'infra
    adapters/                  # implementations concretes
  db/                          # schema SQL
  infra/                       # compose files et bootstrap infra
  build/                       # Dockerfiles
  docs/                        # architecture et ADRs
```

## Fonctionnement attendu

1. L'operateur cree un projet dans l'interface Dockyard.
2. Dockyard recupere le repo et les parametres Docker : branche, dossier racine, Dockerfile, contexte de build.
3. Le worker construit une image immutable et la pousse dans le registry prive.
4. Une release est creee dans le control plane.
5. L'operateur deploye la release vers une cible Docker.
6. Le deploy-agent du serveur applique le deploiement localement.
7. Dockyard configure le routage HTTP(S), a terme via Nginx Proxy Manager.

## Commandes utiles

```bash
make local-infra-up   # demarre l'infra locale de dev
make run-api          # lance l'API
make run-worker       # lance le worker
make run-agent        # lance l'agent
make web-dev          # lance le frontend
make build            # build Go
make test             # tests Go
```

Frontend :

```bash
cd web
npm install
npm run dev
```

## Documentation

- [Architecture cible](./docs/architecture.md)
- [ADR 0001 - Architecture V1](./docs/adrs/0001-architecture-v1.md)
- [Guide architecture et Go](./docs/go-architecture-guide.md)
- [Schema SQL initial CockroachDB](./db/schema.sql)
- [Contrats Go initiaux](./internal/ports/README.md)
- [Infra et deploiement](./infra/README.md)

## Statut

Le projet est en phase d'initialisation. Les fondations Go, Next.js, Docker Compose et les premiers contrats metier sont presents. Les prochaines etapes prioritaires sont la persistence des projets, l'API projet, l'interface de creation de projet, puis le premier workflow build -> release -> deployment.
