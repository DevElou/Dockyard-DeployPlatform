# Redis

Redis est simple en V1 :

- une seule instance
- persistence AOF
- usage pour queue et etat temporaire

Lancement :

```bash
cd infra/foundation/redis
docker compose -f compose.yml up -d
```
