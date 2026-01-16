# Doom Coding Makefile
# =====================
# Build and management tasks for the Doom Coding development environment

.PHONY: all build build-tui install clean help
.PHONY: docker-up docker-down docker-logs docker-build
.PHONY: test lint check health

# Variables
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_CLEAN=$(GO_CMD) clean
GO_MOD=$(GO_CMD) mod

TUI_DIR=cmd/doom-tui
TUI_BINARY=doom-tui
TUI_OUTPUT=bin/$(TUI_BINARY)

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# ============================================================================
# Build targets
# ============================================================================

## build: Build all components
build: build-tui

## build-tui: Build the TUI application
build-tui:
	@echo "Building doom-tui..."
	@mkdir -p bin
	cd $(TUI_DIR) && $(GO_BUILD) $(LDFLAGS) -o ../../$(TUI_OUTPUT) .
	@echo "Built: $(TUI_OUTPUT)"

## build-tui-all: Build TUI for all platforms
build-tui-all:
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	cd $(TUI_DIR) && GOOS=linux GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o ../../bin/$(TUI_BINARY)-linux-amd64 .
	cd $(TUI_DIR) && GOOS=linux GOARCH=arm64 $(GO_BUILD) $(LDFLAGS) -o ../../bin/$(TUI_BINARY)-linux-arm64 .
	cd $(TUI_DIR) && GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o ../../bin/$(TUI_BINARY)-darwin-amd64 .
	cd $(TUI_DIR) && GOOS=darwin GOARCH=arm64 $(GO_BUILD) $(LDFLAGS) -o ../../bin/$(TUI_BINARY)-darwin-arm64 .
	@echo "Built binaries in bin/"

## install: Install TUI to /usr/local/bin
install: build-tui
	@echo "Installing doom-tui to /usr/local/bin..."
	sudo cp $(TUI_OUTPUT) /usr/local/bin/$(TUI_BINARY)
	sudo chmod +x /usr/local/bin/$(TUI_BINARY)
	@echo "Installed: /usr/local/bin/$(TUI_BINARY)"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	cd $(TUI_DIR) && $(GO_CLEAN)
	@echo "Clean complete"

# ============================================================================
# Docker targets
# ============================================================================

## docker-up: Start Docker containers
docker-up:
	@echo "Starting Docker containers..."
	docker compose up -d
	@echo "Containers started"

## docker-up-lxc: Start Docker containers (local network mode)
docker-up-lxc:
	@echo "Starting Docker containers (LXC mode)..."
	docker compose -f docker-compose.lxc.yml up -d
	@echo "Containers started"

## docker-down: Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker compose down
	@echo "Containers stopped"

## docker-logs: View Docker container logs
docker-logs:
	docker compose logs -f

## docker-build: Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker compose build
	@echo "Build complete"

## docker-rebuild: Rebuild Docker images without cache
docker-rebuild:
	@echo "Rebuilding Docker images..."
	docker compose build --no-cache
	@echo "Rebuild complete"

## docker-ps: Show running containers
docker-ps:
	docker compose ps

# ============================================================================
# Development targets
# ============================================================================

## deps: Download Go dependencies
deps:
	@echo "Downloading dependencies..."
	cd $(TUI_DIR) && $(GO_MOD) download
	cd $(TUI_DIR) && $(GO_MOD) tidy
	@echo "Dependencies updated"

## test: Run tests
test:
	@echo "Running tests..."
	cd $(TUI_DIR) && $(GO_TEST) -v ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(TUI_DIR) && golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	cd $(TUI_DIR) && $(GO_CMD) fmt ./...
	@echo "Formatting complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	cd $(TUI_DIR) && $(GO_CMD) vet ./...

# ============================================================================
# Installation targets
# ============================================================================

## setup: Run interactive setup (TUI)
setup: build-tui
	./$(TUI_OUTPUT)

## setup-unattended: Run unattended setup (requires env vars)
setup-unattended:
	./scripts/install.sh --unattended

## setup-local: Run setup for local network (no Tailscale)
setup-local:
	./scripts/install.sh --skip-tailscale

## setup-minimal: Minimal terminal-only setup
setup-minimal:
	./scripts/install.sh --skip-docker --skip-tailscale --skip-hardening --skip-secrets

# ============================================================================
# Utility targets
# ============================================================================

## health: Run health check
health:
	@echo "Running health check..."
	./scripts/health-check.sh

## health-json: Run health check (JSON output)
health-json:
	./scripts/health-check.sh --json

## status: Show system status
status: build-tui
	./$(TUI_OUTPUT) status

## secrets-init: Initialize secrets management
secrets-init:
	./scripts/setup-secrets.sh init

## secrets-edit: Edit encrypted secrets
secrets-edit:
	./scripts/setup-secrets.sh edit secrets/secrets.enc.yaml

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"

# ============================================================================
# Help
# ============================================================================

## help: Show this help message
help:
	@echo ""
	@echo "Doom Coding - Development Environment Setup"
	@echo "============================================"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | grep -E 'build|install|clean' | sed 's/## /  /' | sed 's/: /\t/'
	@echo ""
	@echo "Docker targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | grep -E 'docker' | sed 's/## /  /' | sed 's/: /\t/'
	@echo ""
	@echo "Setup targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | grep -E 'setup|secrets' | sed 's/## /  /' | sed 's/: /\t/'
	@echo ""
	@echo "Development targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | grep -E 'deps|test|lint|fmt|vet' | sed 's/## /  /' | sed 's/: /\t/'
	@echo ""
	@echo "Utility targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | grep -E 'health|status|version' | sed 's/## /  /' | sed 's/: /\t/'
	@echo ""
