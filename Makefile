# Makefile for Web HTTP Exec

# Variables
BINARY_NAME=webhttpexec
BUILD_DIR=build
MAIN_FILE=main.go
PORT?=8080

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
.PHONY: build
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

.PHONY: build-linux
build-linux: ## Build the application for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_FILE)

.PHONY: build-darwin
build-darwin: ## Build the application for macOS
	@echo "Building $(BINARY_NAME) for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin $(MAIN_FILE)

.PHONY: build-windows
build-windows: ## Build the application for Windows
	@echo "Building $(BINARY_NAME) for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_FILE)

.PHONY: build-all
build-all: build build-linux build-darwin build-windows ## Build for all platforms

# Run targets
.PHONY: run
run: ## Run the application
	@echo "Running $(BINARY_NAME) on port $(PORT)..."
	PORT=$(PORT) $(GOCMD) run $(MAIN_FILE)

.PHONY: run-build
run-build: build ## Build and run the application
	@echo "Running $(BINARY_NAME) on port $(PORT)..."
	PORT=$(PORT) ./$(BUILD_DIR)/$(BINARY_NAME)

# Development targets
.PHONY: dev
dev: ## Run in development mode with hot reload (requires air)
	@if command -v air > /dev/null; then \
		echo "Running with air for hot reload..."; \
		air; \
	else \
		echo "Air not found. Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

.PHONY: install-air
install-air: ## Install air for hot reload development
	@echo "Installing air..."
	go install github.com/cosmtrek/air@latest

# Test targets
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-bench
test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. ./...

# Dependency management
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOGET) -v -t -d ./...

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

.PHONY: deps-clean
deps-clean: ## Clean module cache
	@echo "Cleaning module cache..."
	$(GOCLEAN) -modcache

# Code quality targets
.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

.PHONY: vet
vet: ## Vet code
	@echo "Vetting code..."
	$(GOCMD) vet ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: install-lint
install-lint: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Clean targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean deps-clean ## Clean everything including dependencies

# Install targets
.PHONY: install
install: ## Install the application
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_FILE)

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p $(PORT):8080 $(BINARY_NAME)

.PHONY: docker-clean
docker-clean: ## Clean Docker images
	@echo "Cleaning Docker images..."
	docker rmi $(BINARY_NAME) 2>/dev/null || true

# Release targets
.PHONY: release
release: clean build-all ## Build release binaries for all platforms
	@echo "Release binaries created in $(BUILD_DIR)/"

.PHONY: release-linux
release-linux: clean build-linux ## Build release binary for Linux
	@echo "Linux binary created: $(BUILD_DIR)/$(BINARY_UNIX)"

# Utility targets
.PHONY: check
check: fmt vet test ## Run code checks (fmt, vet, test)

.PHONY: pre-commit
pre-commit: fmt vet lint test ## Run all pre-commit checks

.PHONY: setup
setup: deps install-lint install-air ## Setup development environment

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Ensure build directory exists for build targets
build: $(BUILD_DIR)
build-linux: $(BUILD_DIR)
build-darwin: $(BUILD_DIR)
build-windows: $(BUILD_DIR) 