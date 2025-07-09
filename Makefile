# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)

# Setting SHELL to sh for compatibility.
# Options are set to exit when a recipe line exits non-zero.
SHELL = /usr/bin/env sh
.SHELLFLAGS = -ec
GO?=go
GOFLAGS?=

# Default target
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Generate Go code only
.PHONY: generate
generate:
	@echo "Generating Go code..."
	$(GO) generate ./...

# Generate code (including web client)
.PHONY: generate-webclient
generate-webclient:
	@echo "Generating web client..."
	rm -rf web/src/client/ && docker run --rm -v "$$(pwd):/local" openapitools/openapi-generator-cli generate \
		-i /local/api/v1/openapi.yaml \
		-g typescript-fetch \
		-o /local/web/src/client

# Generate API documentation
.PHONY: generate-apidoc
generate-apidoc:
	docker run --rm -v "$$(pwd):/local" openapitools/openapi-generator-cli generate \
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
build: submodule-update deps generate ## Build the application
	@echo "Building application..."
	$(GO) build $(GOFLAGS) -o bin/threadmirror ./cmd/*.go

# Build for Docker (without web client generation)
.PHONY: build-docker
build-docker: submodule-update deps generate ## Build application for Docker
	@echo "Building application for Docker..."
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

# Run tests with testcontainers (default)
.PHONY: test
test: ## Run tests with testcontainers (requires Docker)
	@echo "Running tests with testcontainers..."
	@echo "Note: Docker is required for these tests"
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

# Docker variables
DOCKER_IMAGE_NAME ?= threadmirror
DOCKER_TAG ?= latest
DOCKER_REGISTRY ?= 0x5459

# Build Docker image
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build --platform linux/amd64 -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

# Run Docker container in background
.PHONY: docker-run
docker-run: ## Run Docker container in detached mode
	@echo "Running Docker container in background..."
	docker rm threadmirror-container || true
	docker run  -p 8080:8080 \
		--name threadmirror-container \
		-it --rm \
		--env-file .env \
		$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Push Docker image to registry
.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image to registry..."
	echo "Tagging and pushing to registry: $(DOCKER_REGISTRY)"
	docker tag $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Full Docker workflow
.PHONY: docker-all
docker-all: docker-build docker-run ## Clean, build and run Docker container

# Update git submodules
.PHONY: submodule-update
submodule-update: ## Update git submodules
	@git submodule update --init --recursive
