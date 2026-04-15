GO ?= go
NPM ?= npm

.PHONY: run-api run-worker run-agent fmt build test

run-api:
	$(GO) run ./cmd/control-plane-api

run-worker:
	$(GO) run ./cmd/orchestrator-worker

run-agent:
	$(GO) run ./cmd/deploy-agent

fmt:
	$(GO) fmt ./...

build:
	$(GO) build ./...

test:
	$(GO) test ./...

all: 
	$(MAKE) fmt
	$(MAKE) build
	$(MAKE) test
	$(MAKE) run-api
	$(MAKE) run-worker
	$(MAKE) run-agent
