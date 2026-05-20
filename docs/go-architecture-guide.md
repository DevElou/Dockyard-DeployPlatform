# Dockyard: comprendre l'architecture et Go

Ce document a deux objectifs :

1. t'expliquer comment Dockyard est decoupe
2. te donner juste assez de Go pour etre a l'aise dans ce repo

Le but n'est pas de te faire un cours general sur Go. Le but est de te rendre autonome sur cette architecture.

## 1. Vue d'ensemble de Dockyard

Dockyard est decoupe en quatre zones :

1. `web`
2. `control plane`
3. `agents`
4. `infra shared`

### `web`

Le frontend Next.js sert a :
- afficher l'etat des projets, releases, deployments, domains
- declencher des actions operateur
- consulter l'historique

Le frontend ne parle qu'au `control-plane-api`.

### `control-plane-api`

C'est le cerveau synchrone du systeme.

Il sert a :
- exposer l'API HTTP
- valider les commandes metier
- lire et ecrire l'etat canonique
- publier des jobs pour le worker

Il ne doit pas :
- builder des images lui-meme
- deployer sur les hosts lui-meme
- parler directement au Docker Engine des serveurs cibles

### `orchestrator-worker`

C'est le cerveau asynchrone.

Il sert a :
- executer les builds
- creer les releases
- planifier les deployments
- piloter les rollbacks
- provisionner routing et DNS

Le worker prend des jobs et fait avancer les workflows longs.

### `deploy-agent`

Il tourne sur chaque host Docker cible.

Il sert a :
- recevoir un `DeploymentSpec`
- pull une image
- creer ou remplacer un conteneur
- verifier la sante locale
- remonter un statut

L'agent ne connait pas la logique produit globale. Il execute.

### `infra shared`

Les briques partagees sont :
- CockroachDB : source de verite metier
- Redis : jobs et etat transitoire
- Registry Docker : artefacts
- Nginx Proxy Manager : routage
- DuckDNS : DNS en V1

## 2. Pourquoi cette architecture

Le point important n'est pas "faire microservices". Le point important est d'isoler les responsabilites.

Dockyard repose sur cinq concepts metier :

- `Project`
- `Release`
- `Deployment`
- `RuntimeTarget`
- `Domain`

Le coeur du metier ne doit pas dependre du runtime concret.

Autrement dit :
- aujourd'hui on deploye sur Docker
- demain on pourra deployer sur Kubernetes
- la logique metier ne doit pas etre reecrite

Pour ca, on passe par des interfaces :
- `runtime.Driver`
- `dns.Provider`
- `routing.Provider`
- `source.Provider`
- `registry.Builder`
- `agent.Client`

Le code metier depend des interfaces. Les implementations concretes vivent dans `internal/adapters`.

## 3. Structure du repo

```text
cmd/
  control-plane-api/
  orchestrator-worker/
  deploy-agent/

internal/
  domain/
  application/
  ports/
  adapters/

web/
```

### `cmd/`

`cmd/` contient les points d'entree des binaires.

Chaque dossier sous `cmd/` produit un executable :
- `cmd/control-plane-api`
- `cmd/orchestrator-worker`
- `cmd/deploy-agent`

Dans un projet Go, c'est une convention tres courante.

### `internal/domain`

Ici vivent :
- les types metier
- les enums metier
- les invariants simples

Le domaine ne depend pas de PostgreSQL, Docker, GitHub, Redis ou Nginx Proxy Manager.

Si un fichier du domaine commence a importer un SDK Docker, c'est un mauvais signe.

### `internal/application`

Ici vivent les use cases.

Exemples :
- `CreateProject`
- `CreateDeployment`
- `RollbackDeployment`
- `ListProjects`

L'application orchestre le domaine via des interfaces.

Elle depend de :
- `domain`
- `ports`

Elle ne depend pas directement des implementations concretes.

### `internal/ports`

Les `ports` sont les interfaces du systeme.

Exemples :
- repository SQL
- runtime driver
- dns provider
- queue publisher

Tu peux voir les ports comme les "prises" du coeur applicatif.

### `internal/adapters`

Les `adapters` implementent les ports.

Exemples :
- un repository CockroachDB
- un client DuckDNS
- un driver Docker
- un client HTTP vers l'agent
- un routeur HTTP pour l'API

## 4. Comment lire du Go sans douleur

Go est simple syntactiquement. Le plus important est de comprendre ses conventions.

### Packages

Un dossier = souvent un package.

Exemple :

```go
package domain
```

Tout fichier dans le meme dossier a le meme `package`.

### Export public / prive

En Go, la visibilite depend de la casse :

- `Project` : exporte
- `project` : non exporte

Si un nom commence par une majuscule, il est visible hors du package.

### Structs

Une `struct` est un type compose.

```go
type Project struct {
	ID   string
	Slug string
	Name string
}
```

C'est l'equivalent d'un objet de donnees simple.

### Interfaces

Go utilise beaucoup les interfaces.

```go
type Provider interface {
	EnsureRecord(request RecordRequest) error
}
```

Une interface decrit ce qu'un composant sait faire, pas comment il le fait.

### Erreurs

En Go, les erreurs sont des valeurs retournees.

```go
result, err := service.Do()
if err != nil {
	return err
}
```

Il n'y a pas d'exception comme en JavaScript ou en Java.

### Context

Le `context.Context` sert a porter :
- timeout
- cancellation
- metadata de requete

Tu le verras souvent sur les fonctions d'application, DB et HTTP.

### Receivers

Les methodes sont attachees a un type :

```go
func (s *Service) Execute() error {
	return nil
}
```

Ici `s *Service` est le receiver.

## 5. Le flux d'une requete dans Dockyard

Prenons `POST /api/v1/projects`.

1. la requete HTTP arrive dans `adapters/http`
2. le handler decode le JSON
3. il appelle un use case de `application`
4. le use case valide et utilise un `ProjectRepository`
5. le repository concret ecrit dans CockroachDB
6. le resultat est renvoye en JSON

Autrement dit :

`HTTP -> application -> ports -> adapters concrets`

Pas :

`HTTP handler -> SQL + Docker + Redis en vrac`

## 6. Pourquoi ne pas tout mettre dans `main.go`

Parce que `main.go` doit juste assembler l'application :

- charger la config
- instancier les adapters
- instancier les services
- lancer le serveur

Si `main.go` contient toute la logique, tu obtiens vite un monolithe illisible.

## 7. Le scaffold actuel

Le scaffold que je mets dans ce repo vise trois choses :

1. te donner des binaires clairement separes
2. garder un backend Go compilable sans dependances externes
3. preparer les futures integrations sans detruire la lisibilite

Ce scaffold n'est pas encore le produit final.

Il fournit :
- un `control-plane-api` HTTP minimal
- un `orchestrator-worker` minimal
- un `deploy-agent` minimal
- des interfaces de ports
- des repositories Postgres pour l'etat canonique
- un frontend Next.js minimal

## 8. Ce qui viendra ensuite

L'ordre conseille est :

1. migrations CockroachDB
2. repositories SQL
3. use cases `projects`, `runtime_targets`, `releases`, `deployments`, `domains`
4. agent Docker reel
5. routing Nginx Proxy Manager
6. build pipeline GitHub

## 9. Regles a garder en tete

- `Release` est immutable
- `Deployment` est un evenement metier
- l'agent execute, il n'orchestrationne pas
- le worker orchestre, il ne porte pas l'etat canonique
- le domaine ne depend jamais de Docker
- le frontend ne parle qu'au control plane

## 10. Comment progresser en Go sur ce repo

Pour etre efficace, concentre-toi sur :

1. lire les types dans `internal/domain`
2. lire les interfaces dans `internal/ports`
3. lire le wiring dans `cmd/*/main.go`
4. lire un handler HTTP puis le use case appele

Si tu comprends ce chemin, tu comprends deja la base du projet.
