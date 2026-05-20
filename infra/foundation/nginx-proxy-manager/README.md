# Nginx Proxy Manager

Nginx Proxy Manager sert de point d'entree HTTP(S) pour Dockyard et les applications qu'il gerera.

Avant lancement :

```bash
docker network create dockyard_edge
```

Puis :

```bash
cd infra/foundation/nginx-proxy-manager
docker compose -f compose.yml up -d
```

Ports exposes :

- `80`
- `443`
- `81` pour l'interface d'administration

L'integration Dockyard devra piloter Nginx Proxy Manager via son API pour creer et mettre a jour les proxy hosts.
