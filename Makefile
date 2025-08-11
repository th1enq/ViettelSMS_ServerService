# Variables
PROTO_DIR := proto
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")
GEN_DIR := $(PROTO_DIR)/gen
GO_GEN_DIR := $(GEN_DIR)/go
SWAGGER_DIR := $(GEN_DIR)/swagger
DOCS_DIR := docs

# Go related variables
GO_MODULE := $(shell go list -m)
GO_FILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./$(GEN_DIR)/*")

# Docker and build variables
BINARY_NAME := server-service
BUILD_DIR := build

# Help target
.PHONY: help
help: ## Display this help screen
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Proto generation targets
.PHONY: proto-gen
proto-gen: ## Generate Go code from proto files
	@echo "Generating Go code from proto files..."
	@cd $(PROTO_DIR) && buf generate

.PHONY: swagger-gen
swagger-gen: ## Generate Swagger documentation from proto files
	@echo "Generating Swagger documentation..."
	@mkdir -p $(SWAGGER_DIR)
	@cd $(PROTO_DIR) && buf generate --template buf.gen.swagger.yaml

.PHONY: proto-clean
proto-clean: ## Clean generated proto files
	@echo "Cleaning generated proto files..."
	@rm -rf $(GEN_DIR)

.PHONY: proto-all
proto-all: proto-clean proto-gen swagger-gen ## Clean and regenerate all proto files

# Documentation targets
.PHONY: docs-serve
docs-serve: swagger-gen ## Serve Swagger documentation locally (requires Python)
	@echo "Starting local server for Swagger documentation..."
	@echo "Open http://localhost:8000 in your browser"
	@cd $(SWAGGER_DIR) && python3 -m http.server 8000

.PHONY: docs-serve-docker
docs-serve-docker: swagger-gen ## Serve Swagger documentation using Docker
	@echo "Starting Swagger UI server on http://localhost:8080"
	@docker run -p 8080:8080 -v $(PWD)/$(SWAGGER_DIR):/usr/share/nginx/html/swagger -e SWAGGER_JSON=/swagger/server.swagger.json swaggerapi/swagger-ui

.PHONY: docs-build
docs-build: swagger-gen ## Build documentation
	@echo "Building documentation..."
	@mkdir -p $(DOCS_DIR)
	@cp -r $(SWAGGER_DIR)/* $(DOCS_DIR)/

.PHONY: docs-open
docs-open: swagger-gen ## Open Swagger documentation in browser (requires xdg-open)
	@echo "Opening Swagger documentation in browser..."
	@cd $(SWAGGER_DIR) && python3 -m http.server 8000 > /dev/null 2>&1 & echo $$! > .server.pid
	@sleep 2
	@xdg-open http://localhost:8000 || echo "Please open http://localhost:8000 in your browser"
	@echo "Press Ctrl+C to stop the server"
	@trap 'kill $$(cat $(SWAGGER_DIR)/.server.pid) 2>/dev/null; rm -f $(SWAGGER_DIR)/.server.pid' EXIT; read

# Development targets
.PHONY: install-tools
install-tools: ## Install required development tools
	@echo "Installing development tools..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

.PHONY: tidy
tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

# Build targets
.PHONY: build
build: proto-gen ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

.PHONY: build-migrate
build-migrate: ## Build the migration tool
	@echo "Building migration tool..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/migrate ./cmd/migrate

# Test targets
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Lint and format targets
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

# Run targets
.PHONY: run
run: build ## Run the application
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-dev
run-dev: proto-gen ## Run the application in development mode
	@echo "Running $(BINARY_NAME) in development mode..."
	@go run ./cmd/server

.PHONY: migrate
migrate: build-migrate ## Run database migrations
	@echo "Running database migrations..."
	@./$(BUILD_DIR)/migrate

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(BINARY_NAME):latest

# Clean targets
.PHONY: clean
clean: proto-clean ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DOCS_DIR)
	@rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean ## Clean everything including dependencies
	@echo "Cleaning everything..."
	@go clean -modcache

# Development workflow targets
.PHONY: dev-setup
dev-setup: install-tools deps proto-all ## Set up development environment
	@echo "Development environment setup complete!"

.PHONY: pre-commit
pre-commit: fmt vet lint test proto-all ## Run pre-commit checks
	@echo "Pre-commit checks completed successfully!"

# Default target
.DEFAULT_GOAL := help