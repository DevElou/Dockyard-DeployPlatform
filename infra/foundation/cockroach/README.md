# CockroachDB

Un conteneur par serveur.

## Fichiers

- `server-1.compose.yml`
- `server-2.compose.yml`
- `server-3.compose.yml`

## Variables utiles

- `COCKROACH_ADVERTISE_ADDR`
- `COCKROACH_JOIN`
- `COCKROACH_DATA_DIR`

## Lancement

Sur `server-1` :

```bash
docker compose -f server-1.compose.yml up -d
```

Sur `server-2` :

```bash
docker compose -f server-2.compose.yml up -d
```

Sur `server-3` :

```bash
docker compose -f server-3.compose.yml up -d
```

## Initialisation

Une seule fois, depuis le premier noeud :

```bash
docker exec -it cockroach-1 cockroach init
docker exec -it cockroach-1 cockroach sql -e "CREATE DATABASE dockyard;"
```

Ensuite, tu pourras brancher l'API et le worker sur ce cluster.
