# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
GO?=go

# Default target
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Generate code
.PHONY: generate
generate:
	@echo "Generating code..."
	$(GO) generate ./...
	rm -rf web/src/client/ && docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
		-i /local/api/v1/openapi.yaml \
		-g typescript-fetch \
		-o /local/web/src/client

# Generate API documentation
.PHONY: generate-apidoc
generate-apidoc:
	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
		-i /local/api/v1/openapi.yaml \
		-g html2 \
		-o /local/docs/api/

# Download dependencies
.PHONY: deps
deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Build the application
.PHONY: build
build: deps generate ## Build the application
	@echo "Building application..."
	$(GO) build -o bin/threadmirror ./cmd/*.go

# Run the application
.PHONY: run
run: build ## Run the application
	@echo "Running application..."
	./bin/threadmirror server

# Run development server
.PHONY: dev
dev: ## Run development server
	@echo "Starting development server..."
	$(GO) run ./cmd/*.go  --debug server

# Run tests
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v ./...

# Clean build artifacts
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Lint code
.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	$(GO) tool golangci-lint run

# Lint code and fixing it
.PHONY: lint-fix
lint-fix: ## Run linter and fixing it
	@echo "Running linter and fixing it..."
	$(GO) tool golangci-lint run --fix

# Development setup
.PHONY: setup
setup: deps generate ## Setup development environment
	@echo "Development environment setup complete!"

# Database migration
.PHONY: migrate
migrate: build ## Run database migrations
	@echo "Running database migrations..."
	./bin/threadmirror migrate
