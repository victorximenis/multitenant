# Multitenant Go Library Makefile

.PHONY: help test test-unit test-integration test-coverage lint fmt vet build clean deps docker-up docker-down example

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run

# Testing commands
test: test-unit ## Run all tests

test-unit: ## Run unit tests
	go test -v -race ./...

test-integration: ## Run integration tests (requires Docker services)
	go test -v -race -tags=integration ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-ci: ## Run tests with coverage for CI
	go test -v -race -coverprofile=coverage.out ./...

# Build commands
build: ## Build the library
	go build -v ./...

build-examples: ## Build examples
	cd examples && go build -v ./...

# Docker commands
docker-up: ## Start Docker services for testing
	docker run -d --name postgres-test -p 5432:5432 \
		-e POSTGRES_USER=dev_user \
		-e POSTGRES_PASSWORD=dev_password \
		-e POSTGRES_DB=multitenant_db \
		postgres:15 || true
	docker run -d --name redis-test -p 6379:6379 redis:7 || true
	docker run -d --name mongo-test -p 27017:27017 \
		-e MONGO_INITDB_ROOT_USERNAME=admin \
		-e MONGO_INITDB_ROOT_PASSWORD=password \
		mongo:7 || true
	@echo "Waiting for services to be ready..."
	@sleep 10

docker-down: ## Stop Docker services
	docker stop postgres-test redis-test mongo-test || true
	docker rm postgres-test redis-test mongo-test || true

docker-logs: ## Show Docker services logs
	@echo "=== PostgreSQL logs ==="
	docker logs postgres-test || true
	@echo "=== Redis logs ==="
	docker logs redis-test || true
	@echo "=== MongoDB logs ==="
	docker logs mongo-test || true

# Example commands
example: ## Run the example application
	cd examples && go run main.go

example-build: ## Build the example application
	cd examples && go build -o example main.go

# Quality commands
check: fmt vet lint test ## Run all quality checks

ci: deps check test-coverage-ci ## Run CI pipeline locally

# Cleanup commands
clean: ## Clean build artifacts and test files
	go clean ./...
	rm -f coverage.out coverage.html
	rm -f examples/example

clean-all: clean docker-down ## Clean everything including Docker containers

# Release commands
tag: ## Create a new git tag (usage: make tag VERSION=v0.1.0)
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag VERSION=v0.1.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Documentation commands
docs: ## Generate documentation
	@echo "Generating Go documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"
	@echo "Press Ctrl+C to stop"

# Security commands
security: ## Run security scan
	gosec ./...

# Benchmark commands
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Install tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/github-action-gosec@latest

# Environment setup
setup: install-tools deps ## Setup development environment
	@echo "Development environment setup complete!"
	@echo "Run 'make docker-up' to start test services"
	@echo "Run 'make test' to run tests"

# Git hooks
install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@echo '#!/bin/sh\nmake fmt vet lint' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed!"

# Version info
version: ## Show version information
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD)"
	@echo "Git branch: $(shell git rev-parse --abbrev-ref HEAD)"
	@echo "Build date: $(shell date)" 