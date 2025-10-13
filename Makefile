# Makefile for simple-redirects

# Read Go version from .tool-versions
GO_VERSION := $(shell grep "^golang" .tool-versions | awk '{print $$2}')

# Variables
BINARY_NAME := simple-redirects
BUILD_DIR := dist
CMD_DIR := ./cmd/config-gen
NGINX_VERSION := 1.27.3
IMAGE_NAME := simple-redirects
IMAGE_TAG := latest

# Build flags
LDFLAGS := -ldflags="-w -s"
BUILD_FLAGS := $(LDFLAGS)

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m

.PHONY: help
help: ## Show this help message
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'

.PHONY: all
all: clean fmt vet lint test build ## Run all checks and build

.PHONY: build
build: ## Build the binary to dist/
	@echo "$(COLOR_BOLD)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-linux
build-linux: ## Build for Linux (amd64)
	@echo "$(COLOR_BOLD)Building $(BINARY_NAME) for Linux...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(COLOR_RESET)"

.PHONY: build-darwin
build-darwin: ## Build for macOS (amd64 and arm64)
	@echo "$(COLOR_BOLD)Building $(BINARY_NAME) for macOS...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64$(COLOR_RESET)"

.PHONY: build-all
build-all: build-linux build-darwin ## Build for all platforms

.PHONY: test
test: ## Run tests
	@echo "$(COLOR_BOLD)Running tests...$(COLOR_RESET)"
	@go test -v ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(COLOR_BOLD)Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@go test -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	@go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: $(BUILD_DIR)/coverage.html$(COLOR_RESET)"

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "$(COLOR_BOLD)Running tests with race detection...$(COLOR_RESET)"
	@go test -race -v ./...
	@echo "$(COLOR_GREEN)✓ Race tests passed$(COLOR_RESET)"

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(COLOR_BOLD)Formatting code...$(COLOR_RESET)"
	@gofmt -w -s .
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(COLOR_BOLD)Checking code formatting...$(COLOR_RESET)"
	@test -z "$$(gofmt -l .)" || (echo "$(COLOR_YELLOW)⚠ Files need formatting:$(COLOR_RESET)" && gofmt -l . && exit 1)
	@echo "$(COLOR_GREEN)✓ Code is formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BOLD)Running go vet...$(COLOR_RESET)"
	@go vet ./...
	@echo "$(COLOR_GREEN)✓ go vet passed$(COLOR_RESET)"

.PHONY: lint
lint: ## Run staticcheck
	@echo "$(COLOR_BOLD)Running staticcheck...$(COLOR_RESET)"
	@if ! command -v staticcheck > /dev/null && ! test -f $$(go env GOPATH)/bin/staticcheck; then \
		echo "$(COLOR_YELLOW)⚠ staticcheck not installed, installing...$(COLOR_RESET)"; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		$$(go env GOPATH)/bin/staticcheck ./...; \
	fi
	@echo "$(COLOR_GREEN)✓ staticcheck passed$(COLOR_RESET)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BOLD)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

.PHONY: run
run: build ## Build and run with example DSL
	@echo "$(COLOR_BOLD)Running $(BINARY_NAME) with example DSL...$(COLOR_RESET)"
	@ROUTES_DSL="old.example.com -> https://new.example.com" \
	 OUTPUT_PATH=$(BUILD_DIR)/nginx.conf \
	 TEMPLATE_PATH=cmd/config-gen/nginx.conf.tmpl \
	 $(BUILD_DIR)/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Generated: $(BUILD_DIR)/nginx.conf$(COLOR_RESET)"

.PHONY: docker-custom
docker-custom: ## Build Docker image with custom versions (use: make docker-custom GO_VERSION=1.25.0 NGINX_VERSION=1.26.0)
	@echo "$(COLOR_BOLD)Building Docker image with custom versions...$(COLOR_RESET)"
	@docker build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg NGINX_VERSION=$(NGINX_VERSION) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "$(COLOR_GREEN)✓ Docker image built: $(IMAGE_NAME):$(IMAGE_TAG)$(COLOR_RESET)"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(COLOR_BOLD)Installing development tools...$(COLOR_RESET)"
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "$(COLOR_GREEN)✓ Tools installed$(COLOR_RESET)"

.PHONY: install-hooks
install-hooks: ## Install pre-commit hooks
	@echo "$(COLOR_BOLD)Installing pre-commit hooks...$(COLOR_RESET)"
	@if ! command -v pre-commit > /dev/null; then \
		echo "$(COLOR_YELLOW)⚠ pre-commit not installed. Install with: pip install pre-commit$(COLOR_RESET)"; \
		exit 1; \
	fi
	@pre-commit install
	@pre-commit install --hook-type commit-msg
	@echo "$(COLOR_GREEN)✓ Pre-commit hooks installed$(COLOR_RESET)"

.PHONY: uninstall-hooks
uninstall-hooks: ## Uninstall pre-commit hooks
	@echo "$(COLOR_BOLD)Uninstalling pre-commit hooks...$(COLOR_RESET)"
	@if command -v pre-commit > /dev/null; then \
		pre-commit uninstall; \
		pre-commit uninstall --hook-type commit-msg; \
	fi
	@echo "$(COLOR_GREEN)✓ Pre-commit hooks uninstalled$(COLOR_RESET)"

.PHONY: pre-commit
pre-commit: ## Run pre-commit hooks on all files
	@echo "$(COLOR_BOLD)Running pre-commit hooks on all files...$(COLOR_RESET)"
	@if ! command -v pre-commit > /dev/null; then \
		echo "$(COLOR_YELLOW)⚠ pre-commit not installed. Install with: pip install pre-commit$(COLOR_RESET)"; \
		exit 1; \
	fi
	@pre-commit run --all-files
	@echo "$(COLOR_GREEN)✓ Pre-commit checks passed$(COLOR_RESET)"

.PHONY: version
version: ## Show version information
	@echo "Go version (from .tool-versions): $(GO_VERSION)"
	@echo "Current Go: $$(go version)"
	@echo "Nginx version: $(NGINX_VERSION)"

.PHONY: ci
ci: fmt-check vet lint test ## Run CI checks (fmt-check, vet, lint, test)
	@echo "$(COLOR_GREEN)✓ All CI checks passed$(COLOR_RESET)"

# Default target
.DEFAULT_GOAL := help
