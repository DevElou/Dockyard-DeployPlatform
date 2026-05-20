# Dockyard — Frontend Specification

## Vision

Interface web type Vercel/Render pour un homelab ESXi. L'utilisateur doit pouvoir créer un projet, déclencher un build, suivre le déploiement et gérer les variables d'environnement — le tout sans toucher à la CLI.

L'UI est un SPA qui consomme exclusivement l'API REST du `control-plane-api`.

---

## Stack recommandée

| Choix | Raison |
|---|---|
| **React + TypeScript** | Typage fort, écosystème riche |
| **Vite** | Build rapide, déjà présent dans le repo (`web/`) |
| **TanStack Query** | Gestion cache + polling des statuts async (builds, deploys) |
| **React Router v7** | SPA routing |
| **shadcn/ui + Tailwind** | Composants accessibles, style minimal |

Aucune dépendance côté auth pour V1 (homelab privé, pas d'authentification).

---

## Base URL & conventions

```
Base : http://localhost:8080   (DOCKYARD_API_ADDR)
Content-Type : application/json
```

### Format de réponse

Toutes les réponses suivent la même enveloppe :

```json
// Succès — liste
{ "items": [...] }

// Succès — ressource unique
{ "id": "...", ...champs }

// Erreur
{
  "error": "not_found",
  "message": "project not found"
}
```

### Codes HTTP

| Code | Signification |
|---|---|
| 200 | OK |
| 201 | Ressource créée |
| 204 | Succès sans corps |
| 400 | Validation échouée ou JSON invalide |
| 404 | Ressource introuvable |
| 409 | Conflit (slug/name/version déjà utilisé) |
| 500 | Erreur interne |

---

## Ressources & Endpoints

### Projets

```
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/{id}
DELETE /api/v1/projects/{id}
```

**Créer un projet — corps**
```json
{
  "slug": "my-app",
  "name": "My App",
  "githubOwner": "octocat",
  "githubRepo": "my-app",
  "defaultBranch": "main",
  "rootDirectory": ".",
  "dockerfilePath": "Dockerfile",
  "buildContext": "."
}
```

**Objet Project**
```json
{
  "id": "uuid",
  "slug": "my-app",
  "name": "My App",
  "status": "active",          // "active" | "archived"
  "githubOwner": "octocat",
  "githubRepo": "my-app",
  "defaultBranch": "main",
  "rootDirectory": ".",
  "dockerfilePath": "Dockerfile",
  "buildContext": "."
}
```

---

### Runtime Targets (serveurs Docker)

```
GET   /api/v1/runtime-targets
POST  /api/v1/runtime-targets
GET   /api/v1/runtime-targets/{id}
PATCH /api/v1/runtime-targets/{id}/enable
PATCH /api/v1/runtime-targets/{id}/disable

GET   /api/v1/projects/{id}/runtime-targets
POST  /api/v1/projects/{id}/runtime-targets   { "runtimeTargetId": "uuid" }
```

**Créer un runtime target**
```json
{
  "slug": "homelab-esxi-1",
  "name": "ESXi Node 1",
  "runtimeType": "docker",
  "endpoint": "http://192.168.1.10:9000",
  "agentKeyHash": "<bcrypt hash of agent key>"
}
```

**Objet RuntimeTarget**
```json
{
  "id": "uuid",
  "slug": "homelab-esxi-1",
  "name": "ESXi Node 1",
  "runtimeType": "docker",
  "endpoint": "http://192.168.1.10:9000",
  "enabled": true
}
```

---

### Services de projet (ports & healthcheck)

```
GET  /api/v1/projects/{projectId}/services
POST /api/v1/projects/{projectId}/services
GET  /api/v1/projects/{projectId}/services/{serviceId}
```

**Créer un service**
```json
{
  "name": "web",
  "containerPort": 3000,
  "healthcheckPath": "/healthz",
  "healthcheckPort": 3000,
  "routingEnabled": true
}
```

**Objet ProjectService**
```json
{
  "id": "uuid",
  "projectId": "uuid",
  "name": "web",
  "containerPort": 3000,
  "healthcheckPath": "/healthz",
  "healthcheckPort": 3000,
  "routingEnabled": true
}
```

> Un projet peut avoir plusieurs services (ex: `web` + `worker`). Le premier créé est la valeur par défaut pour les déploiements.

---

### Environnements & variables

```
GET  /api/v1/projects/{projectId}/environments
POST /api/v1/projects/{projectId}/environments

GET  /api/v1/projects/{projectId}/environments/{envId}/variables
PUT  /api/v1/projects/{projectId}/environments/{envId}/variables
DELETE /api/v1/projects/{projectId}/environments/{envId}/variables/{varId}
```

**Créer un environnement**
```json
{ "name": "production" }
```

**Upsert une variable** (PUT — crée ou met à jour par clé)
```json
{
  "key": "DATABASE_URL",
  "value": "postgres://localhost/mydb",
  "isSecret": false
}
```

**Objet EnvironmentVariable**
```json
{
  "id": "uuid",
  "environmentSetId": "uuid",
  "key": "DATABASE_URL",
  "value": "postgres://localhost/mydb",
  "isSecret": false
}
```

> Les variables `isSecret: true` sont stockées de la même façon en V1 (homelab, pas de chiffrement). L'UI devrait masquer leur valeur par défaut (type `password`).

---

### Releases (builds)

```
GET  /api/v1/projects/{projectId}/releases
POST /api/v1/projects/{projectId}/releases
GET  /api/v1/projects/{projectId}/releases/{releaseId}
```

**Créer une release** — déclenche le build en arrière-plan
```json
{
  "version": "v1.2.3",
  "gitRef": "main"
}
```

> Le `gitRef` peut être une branche, un tag ou un SHA complet. Le back résout le commitSHA via GitHub API avant de créer la release.

**Objet Release**
```json
{
  "id": "uuid",
  "projectId": "uuid",
  "version": "v1.2.3",
  "sourceType": "github",
  "gitSha": "abc123...",
  "gitRef": "main",
  "imageRepository": "registry.local/my-app",
  "imageTag": "v1.2.3",
  "imageDigest": "sha256:...",
  "buildStatus": "pending",
  "createdAt": "2026-05-20T10:00:00Z"
}
```

**`buildStatus`** : `pending` → `running` → `succeeded` | `failed`

L'UI doit **poller** `GET /releases/{id}` toutes les 3-5s jusqu'à `succeeded` ou `failed`.

---

### Déploiements

```
GET  /api/v1/projects/{projectId}/deployments
POST /api/v1/projects/{projectId}/deployments
GET  /api/v1/projects/{projectId}/deployments/{deploymentId}
```

**Créer un déploiement** — nécessite une release avec `buildStatus: succeeded`
```json
{
  "releaseId": "uuid",
  "runtimeTargetId": "uuid",
  "strategy": "recreate"
}
```

> `strategy` : `"recreate"` uniquement en V1.

**Objet Deployment**
```json
{
  "id": "uuid",
  "projectId": "uuid",
  "releaseId": "uuid",
  "runtimeTargetId": "uuid",
  "status": "pending",
  "strategy": "recreate",
  "startedAt": null,
  "finishedAt": null,
  "createdAt": "2026-05-20T10:05:00Z"
}
```

**`status`** : `pending` → `deploying` → `healthy` | `failed` | `rolled_back`

L'UI doit **poller** `GET /deployments/{id}` toutes les 3-5s jusqu'à `healthy` ou `failed`.

---

### Domaines

```
GET    /api/v1/projects/{projectId}/domains
POST   /api/v1/projects/{projectId}/domains
GET    /api/v1/projects/{projectId}/domains/{domainId}
DELETE /api/v1/projects/{projectId}/domains/{domainId}
```

**Créer un domaine**
```json
{
  "hostname": "myapp",
  "baseDomain": "home.example.com",
  "provider": "duckdns",
  "routingType": "nginx",
  "tlsEnabled": true
}
```

---

## Flux utilisateur principal

```
Créer projet
  └─ Ajouter un service (containerPort)
  └─ Créer un environnement "production"
       └─ Ajouter les variables (DATABASE_URL, etc.)
  └─ Associer un runtime target

Créer une release (gitRef: "main")
  └─ Poller buildStatus → "succeeded"

Créer un déploiement (releaseId, runtimeTargetId)
  └─ Poller status → "healthy"
```

---

## Pages à implémenter

### 1. `/projects` — Liste des projets
- Tableau : nom, slug, statut, dernière release, dernier déploiement
- Bouton "New Project"

### 2. `/projects/new` — Créer un projet
- Formulaire : slug, name, GitHub owner/repo, branch, dockerfile path

### 3. `/projects/{id}` — Dashboard projet
- Onglets : **Overview** / **Releases** / **Deployments** / **Services** / **Environment** / **Domains** / **Settings**

### 4. `/projects/{id}/releases`
- Liste des releases avec `buildStatus` + badge couleur
- Bouton "New Release" → formulaire `version` + `gitRef`
- Polling automatique des releases `pending`/`running`

### 5. `/projects/{id}/deployments`
- Liste des déploiements avec `status` + badge
- Bouton "Deploy" → sélection release (filtrée sur `buildStatus: succeeded`) + runtime target
- Polling automatique des déploiements actifs

### 6. `/projects/{id}/services`
- Liste des services configurés
- Formulaire ajout : name, containerPort, healthcheckPath, routingEnabled

### 7. `/projects/{id}/environments`
- Sélecteur d'environnement (production, staging…)
- Tableau des variables : clé / valeur masquée si `isSecret` / actions Modifier / Supprimer
- Formulaire inline upsert

### 8. `/runtime-targets` — Gestion des serveurs
- Liste des targets avec statut enabled/disabled
- Formulaire création
- Actions enable/disable

---

## Polling — implémentation recommandée

Utiliser TanStack Query avec `refetchInterval` conditionnel :

```tsx
const { data: release } = useQuery({
  queryKey: ['release', id],
  queryFn: () => fetchRelease(id),
  refetchInterval: (data) =>
    data?.buildStatus === 'pending' || data?.buildStatus === 'running'
      ? 3000
      : false,
})
```

Même pattern pour les déploiements.

---

## Variables d'environnement front

```bash
VITE_API_BASE_URL=http://localhost:8080
```

---

## Gestion des erreurs

Tous les appels API peuvent retourner `{ "error": "...", "message": "..." }`.
Afficher les messages d'erreur directement dans les formulaires pour les codes 400/409, et une toast globale pour les 500.
