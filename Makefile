SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

GO_DIR := implementations/dag-core
SCRIPTS_DIR := scripts
TSDB_URL ?= postgres://dag:dag@timescaledb:5432/dag?sslmode=disable

.PHONY: help dev-up dev-down go-mod-tidy go-mod-download go-fmt go-vet go-test go-build openapi grpc-gen graphql-gen grpc-ci contracts-check sqlc-gen tsdb-schema-lint tsdb-query-lint tsdb-migrate tsdb-seed tsdb-drop-all

help:
	@printf "Targets:\n"
	@printf "  dev-up          Start local stack (observability + app dev)\n"
	@printf "  dev-down        Stop local stack\n"
	@printf "  go-mod-download Download Go modules\n"
	@printf "  go-mod-tidy     Tidy Go modules\n"
	@printf "  go-fmt          Format Go code\n"
	@printf "  go-vet          Run go vet\n"
	@printf "  go-test         Run tests\n"
	@printf "  go-build        Build API binary\n"
	@printf "  openapi         Generate OpenAPI spec\n"
	@printf "  grpc-gen        Generate gRPC code from proto\n"
	@printf "  graphql-gen     Generate GraphQL code from schema\n"
	@printf "  grpc-ci         Run buf lint/breaking for gRPC\n"
	@printf "  contracts-check Regenerate contracts and fail on diff\n"
	@printf "  sqlc-gen        Generate sqlc code from data/tsdb/query\n"
	@printf "  tsdb-schema-lint Validate migrate/seed filenames under data/tsdb/schema\n"
	@printf "  tsdb-query-lint Validate query filenames under data/tsdb/query\n"
	@printf "  tsdb-migrate    Apply SQL in data/tsdb/schema/migrate via migrator container\n"
	@printf "  tsdb-seed       Apply SQL in data/tsdb/schema/seed via migrator container\n"
	@printf "  tsdb-drop-all   Drop all TSDB objects in public schema (requires CONFIRM=YES)\n"

dev-up:
	@bash $(SCRIPTS_DIR)/dev_up.sh

dev-down:
	@bash $(SCRIPTS_DIR)/dev_down.sh

go-mod-download:
	@cd $(GO_DIR) && go mod download

go-mod-tidy:
	@cd $(GO_DIR) && go mod tidy

go-fmt:
	@cd $(GO_DIR) && gofmt -w ./cmd ./internal

go-vet:
	@cd $(GO_DIR) && go vet ./cmd/... ./internal/...

go-test:
	@cd $(GO_DIR) && go test ./cmd/... ./internal/...

go-build:
	@cd $(GO_DIR) && go build -trimpath -o bin/api ./cmd/api

openapi:
	@mkdir -p implementations/dag-core/openapi
	@cd $(GO_DIR) && go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g main.go -d ./cmd/api,./internal/interface/http -o ./openapi

grpc-gen:
	@mkdir -p $(GO_DIR)/internal/interface/grpc/gen
	@cd $(GO_DIR)/internal/interface/grpc && buf generate

graphql-gen:
	@echo "graphql-gen is skipped: GraphQL transport is currently placeholder."

grpc-ci:
	@cd $(GO_DIR)/internal/interface/grpc && buf lint
	@cd $(GO_DIR)/internal/interface/grpc && buf breaking --against .

contracts-check:
	@$(MAKE) openapi
	@$(MAKE) grpc-gen
	@$(MAKE) graphql-gen
	@git diff --exit-code implementations/dag-core/openapi $(GO_DIR)/internal/interface/grpc/gen

sqlc-gen:
	@cd $(GO_DIR)/internal/infrastructure/persistence/tsdb/query && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate -f sqlc.yaml

tsdb-schema-lint:
	@bash $(SCRIPTS_DIR)/tsdb/lint_migrations.sh

tsdb-query-lint:
	@bash $(SCRIPTS_DIR)/tsdb/lint_queries.sh

tsdb-migrate:
	@TSDB_URL="$(TSDB_URL)" docker compose -f deployments/compose/docker-compose.migrator.yml run --rm tsdb-migrator

tsdb-seed:
	@TSDB_URL="$(TSDB_URL)" MIGRATIONS_DIR=data/tsdb/schema/seed docker compose -f deployments/compose/docker-compose.migrator.yml run --rm -e MIGRATIONS_DIR tsdb-migrator

tsdb-drop-all:
	@if [ "$(CONFIRM)" != "YES" ]; then \
		echo "Refusing to drop DB objects. Re-run with: make tsdb-drop-all CONFIRM=YES"; \
		exit 1; \
	fi
	@docker exec compose-timescaledb-1 psql -U dag -d dag -v ON_ERROR_STOP=1 -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"
