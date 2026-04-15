# Dockyard Architecture

## Vue d'ensemble

Dockyard est structuré autour de quatre zones techniques clairement séparées :

1. `control plane`
2. `deployment agents`
3. `edge / routing`
4. `persistence`

La V1 est `Docker-first`, mais la logique metier ne depend pas directement de Docker. Le coeur du systeme repose sur les ressources `Project`, `Release`, `Deployment`, `RuntimeTarget` et `Domain`.

## Composants

```text
                                   +----------------------+
                                   |      Next.js Web     |
                                   |  UI privee operateur |
                                   +----------+-----------+
                                              |
                                              v
                                   +----------------------+
                                   |  Control Plane API   |
                                   |      Go / chi        |
                                   +----+-----------+-----+
                                        |           |
                                        |           |
                                        v           v
                              +----------------+  +----------------+
                              |  CockroachDB   |  | Redis / BullMQ |
                              | source of truth|  | jobs / state   |
                              +----------------+  +--------+-------+
                                                          |
                                                          v
                                              +----------------------+
                                              | Orchestrator Worker  |
                                              |     Go / asynq       |
                                              +----+-----------+-----+
                                                   |           |
                                                   |           |
                                                   v           v
                                          +-------------+   +------------------+
                                          |  Registry   |   | GitHub Provider  |
                                          |   Docker    |   | webhooks / clone |
                                          +-------------+   +------------------+
                                                   |
                                                   v
                                         +----------------------+
                                         | Deploy Agent (host)  |
                                         |         Go           |
                                         +----------+-----------+
                                                    |
                                                    v
                                               +---------+
                                               | Traefik |
                                               +----+----+
                                                    |
                                                    v
                                                +--------+
                                                | DNS    |
                                                |DuckDNS |
                                                +--------+
```

## Services

### `web`

Responsabilites :
- afficher l'etat des projets, releases, deployments, domains
- declencher les actions operateur
- consulter logs et historique

Stack :
- Next.js
- auth par session simple

### `control-plane-api`

Responsabilites :
- exposer l'API metier
- gerer auth et autorisation simple
- maintenir l'etat canonique
- valider les transitions metier
- publier des jobs asynchrones

Stack :
- Go
- chi
- pgx
- sqlc

### `orchestrator-worker`

Responsabilites :
- executer les workflows asynchrones
- construire les images
- creer les releases immutables
- calculer et appliquer les plans de deploiement
- piloter rollback et provisionnement de domaine

Stack :
- Go
- asynq

### `deploy-agent`

Responsabilites :
- executer localement le plan de deploiement sur un host
- tirer l'image, demarrer, remplacer, verifier
- remonter les evenements et statuts

Regle :
- agent simple, execution locale uniquement
- aucune orchestration globale dans l'agent

### `edge`

Responsabilites :
- routage HTTP(S)
- terminaison TLS
- exposition domaine -> service

Choix V1 :
- Traefik dedie a Dockyard
- source de verite routee depuis Dockyard

### `persistence`

CockroachDB :
- etat metier
- historique
- audit minimal

Redis :
- queue
- verrous courts
- etat temporaire

## Flux metier

### Connexion GitHub

1. creation d'un `GitHubInstallation`
2. association du repo a un `Project`
3. stockage du `default_branch`, `root_directory`, `dockerfile_path`

### Build

1. un `BuildJob` est cree
2. le worker recupere le commit SHA
3. construction de l'image avec `docker buildx build --push`
4. stockage dans le registry prive
5. creation d'une `Release`

### Release

Une `Release` contient :
- source Git resolue
- image repository
- image tag immutable
- image digest
- metadata de build

Une release n'est jamais modifiee apres creation.

### Deployment

1. selection d'une release et d'un runtime target
2. creation d'un `Deployment`
3. le worker construit un `RuntimeDeploymentSpec`
4. le spec est envoye a l'agent
5. l'agent applique le deploiement
6. les statuts remontent dans le control plane

### Domaine

1. creation d'une ressource `Domain`
2. provision DNS via `DnsProvider`
3. configuration Traefik via `RoutingProvider`
4. health check externe eventuel

### Rollback

1. selection d'une release precedente
2. creation d'un nouveau deployment reference comme rollback
3. reapplication de l'image precedente

Le rollback cree un nouvel evenement metier ; il ne remplace pas l'historique.

## Frontieres de conception

Le metier parle a des abstractions :

- `SourceProvider`
- `Builder`
- `Registry`
- `RuntimeDriver`
- `DnsProvider`
- `RoutingProvider`
- `AgentClient`

La V1 n'implemente qu'un seul backend concret pour la plupart de ces abstractions :

- `GitHubSourceProvider`
- `DockerBuildxBuilder`
- `DockerRegistry`
- `DockerRuntimeDriver`
- `DuckDnsProvider`
- `TraefikRoutingProvider`

## Structure Go retenue

```text
dockyard/
  web/
  cmd/
    control-plane-api/
      main.go
    orchestrator-worker/
      main.go
    deploy-agent/
      main.go
  internal/
    domain/
      project/
      release/
      deployment/
      runtimetarget/
      domainroute/
    application/
      command/
      query/
      service/
    ports/
      repository/
      runtime/
      source/
      dns/
      routing/
      queue/
      registry/
      agent/
    adapters/
      http/
      postgres/
      redis/
      github/
      docker/
      traefik/
      duckdns/
      agentclient/
  db/
  docs/
```

## Regles de structure

- `cmd/` contient les points d'entree des binaires.
- `internal/domain` contient les types et regles metier pures.
- `internal/application` contient les use cases.
- `internal/ports` contient les interfaces.
- `internal/adapters` contient les implementations concretes.

Le domaine ne depend d'aucun SDK externe. Docker, GitHub, Traefik et DuckDNS restent strictement dans `adapters`.

## Ce qu'il faut coder d'abord

Premier slice vertical recommande :

1. modele SQL
2. CRUD `Project`, `RuntimeTarget`, `Release`, `Deployment`, `Domain`
3. agent minimal
4. deployment manuel d'une image existante
5. routage Traefik
6. ensuite seulement l'import GitHub et le build
