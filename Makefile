SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

GO_DIR := dag-core
SCRIPTS_DIR := scripts

.PHONY: help dev-up dev-down go-mod-tidy go-mod-download go-fmt go-vet go-test go-build openapi-go openapi-ts openapi-gen openapi-check openapi-lint openapi-breaking

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
	@printf "  openapi-go      Generate Go types/handlers from OpenAPI\n"
	@printf "  openapi-ts      Generate TS types/hooks from OpenAPI\n"
	@printf "  openapi-gen     Generate all OpenAPI artifacts\n"
	@printf "  openapi-check   Fail if OpenAPI generated files differ\n"
	@printf "  openapi-lint    Lint OpenAPI spec with Spectral\n"
	@printf "  openapi-breaking Fail on OpenAPI breaking changes\n"

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

openapi-go:
	@mkdir -p openapi/dag-core/gen/go
	@go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		-config openapi/dag-core/oapi-codegen.yaml \
		openapi/dag-core/openapi.yaml

openapi-ts:
	@mkdir -p openapi/dag-core/gen/ts
	@npx -y orval@6.31.0 --config openapi/dag-core/orval.config.ts

openapi-gen: openapi-go openapi-ts

openapi-check:
	@$(MAKE) openapi-gen
	@git diff --exit-code -- openapi/dag-core/gen

openapi-lint:
	@NPM_CONFIG_CACHE=/tmp/npm-cache npx -y @stoplight/spectral-cli@6.14.0 lint \
		-r openapi/.spectral.yaml \
		openapi/dag-core/openapi.yaml

openapi-breaking:
	@bash scripts/openapi_breaking_check.sh
