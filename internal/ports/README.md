# Ports

Les ports definissent les interfaces entre le coeur metier Dockyard et les implementations concretes.

Regle :
- `internal/domain` ne depend d'aucun port
- `internal/application` depend des ports
- `internal/adapters` implemente les ports

Ports de V1 :
- `repository`
- `runtime`
- `source`
- `dns`
- `routing`
- `queue`
- `registry`
- `agent`
