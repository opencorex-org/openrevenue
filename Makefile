.DEFAULT_GOAL := help
.PHONY: help setup dev api worker scheduler web test lint format migrate seed generate docker-up docker-down
help: ## Show commands
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
setup: ## Download Go and frontend dependencies
	go mod download
	pnpm install
dev: docker-up ## Start local dependencies and API
	go run ./apps/api/cmd/api
api: ## Run API
	go run ./apps/api/cmd/api
worker: ## Run worker
	go run ./apps/worker/cmd/worker
scheduler: ## Run scheduler
	go run ./apps/scheduler/cmd/scheduler
web: ## Run frontend workspace
	pnpm dev
test: ## Run backend and frontend tests
	go test ./...
	pnpm test
lint: ## Run backend and frontend checks
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))"
	go vet ./...
	pnpm lint
format: ## Format sources
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')
	pnpm format
migrate: ## Apply database migrations
	migrate -path database/migrations -database "$${DATABASE_URL:-postgres://openrevenue:openrevenue_dev_only@localhost:5432/openrevenue?sslmode=disable}" up
seed: ## Load fictional development seed data
	psql "$${DATABASE_URL}" -f database/seeds/development.sql
generate: ## Generate contract and query clients
	@echo "Run oapi-codegen, sqlc, and the TypeScript OpenAPI generator in tools/codegen"
docker-up: ## Start local services
	docker compose up -d
docker-down: ## Stop local services
	docker compose down
