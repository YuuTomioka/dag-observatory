SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

GO_DIR := demo-go
SCRIPTS_DIR := scripts

.PHONY: help dev-up dev-down go-mod-tidy go-mod-download go-fmt go-vet go-test go-build

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
