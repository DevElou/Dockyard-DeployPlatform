# Traefik

Traefik sert de point d'entree HTTP(S) pour Dockyard et les applications qu'il gerera.

Avant lancement :

```bash
docker network create dockyard_edge
```

Puis :

```bash
cd infra/foundation/traefik
docker compose -f compose.yml up -d
```

Ports exposes :

- `80`
- `443`
- `8081` pour le dashboard local
