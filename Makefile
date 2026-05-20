GO ?= go
NPM ?= npm
DOCKER_COMPOSE ?= docker-compose
GOCACHE ?= $(CURDIR)/.gocache
DATABASE_URL ?= postgresql://root@localhost:26257/dockyard?sslmode=disable
MIGRATE_URL ?= cockroachdb://root@localhost:26257/dockyard?sslmode=disable
MIGRATE ?= migrate

.PHONY: run-api run-worker run-agent web-dev fmt build test local-infra-up local-infra-down local-infra-logs all db-create migrate-up migrate-down migrate-new test-integration

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
