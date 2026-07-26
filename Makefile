.DEFAULT_GOAL := help

.PHONY: help bootstrap setup doctor validate-config dev api worker scheduler web \
	services-up services-down services-logs services-status go-format go-format-check \
	go-lint go-test web-install web-format web-format-check web-lint web-typecheck \
	web-test web-build format format-check lint typecheck test build quality migrate seed generate \
	contracts workflow-policy security-baseline

help: ## Show supported developer commands
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

bootstrap: ## Validate tools, create .env, and install pinned dependencies
	@./scripts/dev/bootstrap.sh

setup: bootstrap ## Backwards-compatible alias for bootstrap

doctor: ## Diagnose local prerequisite and version problems
	@./scripts/dev/doctor.sh

validate-config: ## Validate pinned versions and safe example configuration
	@./scripts/dev/validate-config.sh

dev: services-up ## Start local services, then run the API
	@$(MAKE) api

api: ## Run the API
	go run ./apps/api/cmd/api

worker: ## Run the worker
	go run ./apps/worker/cmd/worker

scheduler: ## Run the scheduler
	go run ./apps/scheduler/cmd/scheduler

web: ## Run the frontend workspace
	corepack pnpm dev

services-up: ## Start the pinned local stack and wait for healthy dependencies
	docker compose up -d --wait

services-down: ## Stop the local stack without deleting developer data
	docker compose down

services-logs: ## Follow local stack logs
	docker compose logs --follow --tail=100

services-status: ## Show local stack state and health
	docker compose ps

go-format: ## Format Go source
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

go-format-check: ## Check Go formatting without modifying files
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || \
		{ echo "Go files need formatting; run 'make go-format'." >&2; exit 1; }

go-lint: ## Run Go static analysis
	go vet ./...

go-test: ## Run Go tests with race detection
	go test -race -cover ./...

web-install: ## Install exact frontend dependencies
	corepack pnpm install --frozen-lockfile

web-format: ## Format frontend TypeScript sources
	corepack pnpm format

web-format-check: ## Check frontend TypeScript formatting
	corepack pnpm format:check

web-lint: ## Run frontend lint tasks
	corepack pnpm lint

web-typecheck: ## Run frontend type checks
	corepack pnpm typecheck

web-test: ## Run frontend tests
	corepack pnpm test

web-build: ## Build all frontend workspaces
	corepack pnpm build

contracts: ## Validate OpenAPI and event JSON Schema contracts
	corepack pnpm contracts

workflow-policy: ## Validate GitHub Actions syntax and security policy
	@go tool actionlint
	@bash scripts/ci/validate-workflows.sh

security-baseline: ## Validate secure configuration and exception policy
	@bash scripts/ci/validate-security-baseline.sh

format: go-format web-format ## Format all supported source

format-check: go-format-check web-format-check ## Check formatting exactly as CI does

lint: go-lint web-lint ## Run backend and frontend lint

typecheck: web-typecheck ## Run static type checking

test: go-test web-test ## Run backend and frontend tests

build: web-build ## Build deployable artifacts
	go build ./...

quality: validate-config workflow-policy security-baseline format-check lint typecheck test contracts build ## Run the complete local/CI quality baseline

migrate: ## Apply database migrations
	migrate -path database/migrations -database "$${DATABASE_URL:-postgres://openrevenue:openrevenue_dev_only@localhost:5432/openrevenue?sslmode=disable}" up

seed: ## Load fictional development seed data
	@test -n "$${DATABASE_URL:-}" || { echo "DATABASE_URL is required; source .env first." >&2; exit 1; }
	psql "$${DATABASE_URL}" -f database/seeds/development.sql

generate: ## Generate reproducible contract clients
	corepack pnpm contracts:generate
