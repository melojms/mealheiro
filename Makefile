COMPOSE ?= docker compose
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: up up-proxy pull-up down logs build test test-go test-web generate fmt dev dev-api dev-web web e2e

up: ## Build and start (publishes $$HOST_PORT)
	mkdir -p $${DATA_LOCATION:-./data}
	VERSION=$(VERSION) $(COMPOSE) up -d --build --force-recreate

up-proxy: ## Build and start on the external "proxy" network (homelab)
	VERSION=$(VERSION) $(COMPOSE) -f compose.yaml -f compose.proxy.yaml up -d --build --force-recreate

pull-up: ## Pull the released image from GHCR and start (no local build)
	mkdir -p $${DATA_LOCATION:-./data}
	$(COMPOSE) pull
	$(COMPOSE) up -d --no-build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

build: web
	CGO_ENABLED=0 go build -ldflags "-X github.com/melojms/mealheiro/internal/api.Version=$(VERSION)" -o bin/mealheiro ./cmd/mealheiro

web:
	cd web && npm ci && npm run build

test: test-go test-web

test-go:
	go test -race ./...

test-web:
	cd web && npx vitest run

generate:
	sqlc generate

fmt:
	gofumpt -w .

dev-api: ## Go API on :8080 with data in ./data
	DATA_DIR=./data LOG_LEVEL=debug go run ./cmd/mealheiro

dev-web: ## Vite dev server (proxies /api to :8080)
	cd web && npm run dev

e2e:
	cd e2e && npx playwright test
