SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

GO_DIR := dag-core
SCRIPTS_DIR := scripts

.PHONY: help dev-up dev-down go-mod-tidy go-mod-download go-fmt go-vet go-test go-build openapi grpc-gen graphql-gen grpc-ci contracts-check sqlc-gen

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
	@printf "  sqlc-gen        Generate sqlc code from docs/tsdb/query\n"

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
	@cd $(GO_DIR) && go vet ./...

go-test:
	@cd $(GO_DIR) && go test ./...

go-build:
	@cd $(GO_DIR) && go build -trimpath -o bin/api ./cmd/api

openapi:
	@mkdir -p docs/openapi
	@cd $(GO_DIR) && go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g main.go -d ./cmd/api,./internal/interface/http -o ../docs/openapi

grpc-gen:
	@mkdir -p $(GO_DIR)/internal/interface/grpc/gen
	@cd $(GO_DIR) && buf generate --path internal/interface/grpc/proto

graphql-gen:
	@mkdir -p $(GO_DIR)/internal/interface/graphql/schema
	@cd $(GO_DIR) && go run github.com/99designs/gqlgen@v0.17.60 generate -c internal/interface/graphql/gqlgen.yml

grpc-ci:
	@cd $(GO_DIR)/internal/interface/grpc && buf lint
	@cd $(GO_DIR)/internal/interface/grpc && buf breaking --against .

contracts-check:
	@$(MAKE) openapi
	@$(MAKE) grpc-gen
	@$(MAKE) graphql-gen
	@git diff --exit-code docs/openapi $(GO_DIR)/internal/interface/grpc/gen $(GO_DIR)/internal/interface/graphql/gen

sqlc-gen:
	@cd $(GO_DIR) && sqlc generate -f internal/infrastructure/persistence/tsdb/query/sqlc.yaml
