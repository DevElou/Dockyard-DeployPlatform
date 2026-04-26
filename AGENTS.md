# Repository Guidelines

## Project Structure & Module Organization

Dockyard is a Docker-first deployment platform with a Go backend and a Next.js frontend. Go entrypoints live in `cmd/control-plane-api`, `cmd/orchestrator-worker`, and `cmd/deploy-agent`. Core backend code is organized by clean architecture layers under `internal/domain`, `internal/application`, `internal/ports`, and `internal/adapters`.

The frontend lives in `web/`, with app routes and styles under `web/app/`. Dockerfiles are in `build/`, local and service Compose stacks are in `infra/`, the initial CockroachDB schema is in `db/schema.sql`, and architecture notes are in `docs/`.

## Build, Test, and Development Commands

- `make local-infra-up`: starts the local Docker Compose infrastructure from `infra/local/docker-compose.yml`.
- `make run-api`, `make run-worker`, `make run-agent`: run the Go services locally.
- `make web-dev`: starts the Next.js development server from `web/`.
- `make fmt`: runs `go fmt ./...`.
- `make build`: builds all Go packages.
- `make test`: runs all Go tests.
- `make all`: runs format, build, and test.

For frontend setup, run `cd web && npm install`. Useful scripts are `npm run dev`, `npm run build`, `npm run start`, and `npm run lint`.

## Coding Style & Naming Conventions

Use standard Go formatting via `go fmt`; keep package names short, lowercase, and domain-oriented. Exported Go identifiers should be named for their role, such as `ProjectRepository` or `CreateProjectInput`.

For TypeScript and React, use PascalCase for components and camelCase for values. Keep CSS class names descriptive, for example `page-shell`, `hero`, and `card`.

## Testing Guidelines

There are currently no committed test files. Add Go tests next to the package under test using `*_test.go`, then run `make test` or `go test ./...`. Prefer table-driven tests for domain validation and application service behavior.

When frontend behavior becomes stateful, add tests under `web/` using the chosen framework. Until then, verify frontend changes with `cd web && npm run build` and `npm run lint`.

## Commit & Pull Request Guidelines

Existing history uses short conventional-style messages such as `init(project): infra` and `init(project): doc`. Continue with concise, scoped messages like `feat(api): add project creation` or `docs(infra): clarify local setup`.

Pull requests should include a summary, validation commands, linked issues when applicable, and screenshots for visible UI changes. Keep PRs focused on one backend, frontend, infra, or docs concern.

## Security & Configuration Tips

Do not commit secrets, tokens, registry credentials, or local environment files. Keep service configuration in Compose files under `infra/`, and document required environment variables in README or service-specific infra docs.
