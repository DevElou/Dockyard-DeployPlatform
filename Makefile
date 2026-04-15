GO ?= go
NPM ?= npm
DOCKER_COMPOSE ?= docker-compose
GOCACHE ?= $(CURDIR)/.gocache

.PHONY: run-api run-worker run-agent web-dev fmt build test local-infra-up local-infra-down local-infra-logs all

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

all: 
	$(MAKE) fmt
	$(MAKE) build
	$(MAKE) test
