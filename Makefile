GO ?= go
NPM ?= npm
DOCKER_COMPOSE ?= docker-compose
GOCACHE ?= $(CURDIR)/.gocache
DATABASE_URL ?= postgresql://root@localhost:26257/dockyard?sslmode=disable
MIGRATE_URL ?= cockroachdb://root@localhost:26257/dockyard?sslmode=disable
MIGRATE ?= migrate
DEV_COMPOSE = $(DOCKER_COMPOSE) -f docker-compose.dev.yml --env-file .env.dev
PLATFORM_COMPOSE = cd infra/platform/dockyard && $(DOCKER_COMPOSE) --env-file ../../../.env -f compose.yml
AGENT_COMPOSE = cd infra/agents/deploy-agent && $(DOCKER_COMPOSE) --env-file ../../../.env -f compose.yml

.PHONY: run-api run-worker run-agent web-dev fmt build test \
        local-infra-up local-infra-down local-infra-logs \
        dev-up dev-down dev-logs dev-build dev-deploy-web dev-ps \
        deploy deploy-platform deploy-web deploy-agent deploy-ps \
        db-create migrate-up migrate-down migrate-new test-integration all

run-api:
	env GOCACHE=$(GOCACHE) $(GO) run ./cmd/control-plane-api

run-worker:
	env GOCACHE=$(GOCACHE) $(GO) run ./cmd/orchestrator-worker

run-agent:
	env GOCACHE=$(GOCACHE) $(GO) run ./cmd/deploy-agent

web-dev:
	cd web && $(NPM) run dev

fmt:
	$(GO) fmt ./...

build:
	env GOCACHE=$(GOCACHE) $(GO) build ./...

test:
	env GOCACHE=$(GOCACHE) $(GO) test ./...

local-infra-up:
	$(DOCKER_COMPOSE) -f infra/local/docker-compose.yml up -d

local-infra-down:
	$(DOCKER_COMPOSE) -f infra/local/docker-compose.yml down

local-infra-logs:
	$(DOCKER_COMPOSE) -f infra/local/docker-compose.yml logs -f

## ── Stack locale complète (Docker Compose) ─────────────────────────────────

## Démarrer toute la stack (build si nécessaire)
dev-up:
	$(DEV_COMPOSE) up -d

## Démarrer avec rebuild forcé de toutes les images
dev-build:
	$(DEV_COMPOSE) up -d --build

## Rebuild + recréer uniquement le frontend local
dev-deploy-web:
	$(DEV_COMPOSE) up -d --build --no-deps web

## Arrêter la stack (conserver les volumes)
dev-down:
	$(DEV_COMPOSE) down

## Arrêter la stack et supprimer les volumes (reset complet)
dev-reset:
	$(DEV_COMPOSE) down -v

## Suivre les logs de tous les services
dev-logs:
	$(DEV_COMPOSE) logs -f

## Afficher l'état des containers
dev-ps:
	$(DEV_COMPOSE) ps

## ── Déploiement plateforme ─────────────────────────────────────────────────

## Déployer/rebuilder la plateforme principale (API, worker, web)
deploy: deploy-platform

deploy-platform:
	$(PLATFORM_COMPOSE) up -d --build

## Rebuild + recréer uniquement le frontend de la plateforme
deploy-web:
	$(PLATFORM_COMPOSE) up -d --build --no-deps web

## Déployer/rebuilder l'agent local au host cible
deploy-agent:
	$(AGENT_COMPOSE) up -d --build

deploy-ps:
	$(PLATFORM_COMPOSE) ps

db-create:
	cockroach sql --insecure --host=localhost:26257 -e "CREATE DATABASE IF NOT EXISTS dockyard;"

migrate-up:
	$(MIGRATE) -path db/migrations -database "$(MIGRATE_URL)" up

migrate-down:
	$(MIGRATE) -path db/migrations -database "$(MIGRATE_URL)" down 1

migrate-new:
	$(MIGRATE) create -ext sql -dir db/migrations -seq $(name)

test-integration:
	env GOCACHE=$(GOCACHE) DOCKYARD_TEST_DSN=$(DATABASE_URL) $(GO) test -race -tags=integration ./internal/adapters/postgres/...

all:
	$(MAKE) fmt
	$(MAKE) build
	$(MAKE) test
