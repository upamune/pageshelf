.DEFAULT_GOAL := help

.PHONY: help fmt lint test ci

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*## "; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format Go code.
	golangci-lint fmt

lint: ## Run linters.
	golangci-lint run ./...

test: ## Run tests.
	go test ./...

ci: ## Run format check, lint, and tests.
	golangci-lint fmt --diff
	golangci-lint run ./...
	go test ./...
