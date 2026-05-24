# gibson-tool-runner — one microVM image, one Go binary, N parsers.
#
# Targets:
#   make build      Org-contract alias for bin (gibson#171 slice 1.4)
#   make bin        Build the runner binary to ./bin/gibson-runner
#   make test       Run unit tests (parsers, registry)
#   make check      CI-equivalent gate: lint + test (org contract)
#   make list-tools Build the binary and print its catalog
#   make lint       go vet + staticcheck (when available)
#   make image      Build the OCI image for local smoke via Setec
#   make clean      Remove ./bin and ./out

SHELL := /usr/bin/env bash
.SHELLFLAGS := -eo pipefail -c

BIN_DIR := bin
IMAGE   ?= ghcr.io/zero-day-ai/gibson-tool-runner:dev

.PHONY: help
help: ## List targets.
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# build: org Makefile contract target (gibson#171 slice 1.4 /
# zero-day-ai/.github#87). Aliases bin so CI and the drift-detector
# find a canonical build target.
.PHONY: build
build: bin ## Build the runner binary (org-contract alias for bin).

.PHONY: bin
bin: ## Build ./bin/gibson-runner (CGO disabled, static).
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/gibson-runner ./cmd/gibson-runner

.PHONY: test
test: ## Run unit tests with the race detector.
	go test -race ./...

.PHONY: test-coverage
test-coverage: ## Produce a coverage report.
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: list-tools
list-tools: bin ## Build the binary and print its parser catalog as JSON.
	./$(BIN_DIR)/gibson-runner --list-tools

.PHONY: lint
lint: ## Run go vet (staticcheck if installed).
	go vet ./...
	@if command -v staticcheck >/dev/null; then staticcheck ./... ; else echo "staticcheck not installed; skipping"; fi

# check: org Makefile contract CI-equivalent gate (gibson#171 slice 1.4 /
# zero-day-ai/.github#87). Runs lint + test, matching what CI runs on
# every PR.
.PHONY: check
check: lint test ## Run the full CI gate locally (lint + test).

.PHONY: image
image: ## Build the runner OCI image.
	docker build -t $(IMAGE) .

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf $(BIN_DIR) coverage.out coverage.html

.PHONY: tidy
tidy: ## go mod tidy.
	go mod tidy
