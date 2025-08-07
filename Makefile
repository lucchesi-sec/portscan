# Go TUI Port Scanner - Makefile
# High-performance terminal-based port scanner

.PHONY: help build test lint clean install dev dev-setup benchmark test-coverage fuzz security release docker

# Default target
.DEFAULT_GOAL := help

# Build variables
BINARY_NAME := portscan
MAIN_PATH := ./cmd/main.go
BUILD_DIR := ./bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME) -s -w"

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := gofumpt
GOLINT := golangci-lint

## help: Show this help message
help:
	@echo "Go TUI Port Scanner - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build binaries for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ All platform binaries built in $(BUILD_DIR)/"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)
	@echo "✅ $(BINARY_NAME) installed to $(shell go env GOPATH)/bin"

## test: Run all tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...
	@echo "✅ All tests passed"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@$(GOCMD) tool cover -func=coverage.out | tail -1

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@$(GOTEST) -v -race ./...
	@echo "✅ Race tests passed"

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./internal/core -run=^$$
	@echo "✅ Benchmarks completed"

## fuzz: Run fuzz tests
fuzz:
	@echo "Running fuzz tests..."
	@$(GOTEST) -fuzz=FuzzPortParser -fuzztime=30s ./pkg/parser || true
	@echo "✅ Fuzz tests completed"

## lint: Run linters
lint:
	@echo "Running linters..."
	@$(GOLINT) run ./...
	@echo "✅ Linting passed"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -w .
	@goimports -w .
	@echo "✅ Code formatted"

## security: Run security checks
security:
	@echo "Running security checks..."
	@govulncheck ./...
	@gosec ./...
	@echo "✅ Security checks passed"

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@$(GOCMD) clean -cache -testcache -modcache
	@echo "✅ Cleaned"

## deps: Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) verify
	@$(GOMOD) tidy
	@echo "✅ Dependencies updated"

## dev-setup: Setup development environment
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@$(GOCMD) install mvdan.cc/gofumpt@latest
	@$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.2.2
	@$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest
	@$(GOCMD) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GOCMD) install github.com/goreleaser/goreleaser@latest
	@echo "✅ Development environment ready"

## dev: Run in development mode with hot reload
dev:
	@echo "Starting development mode..."
	@echo "Usage: make dev TARGET=localhost PORTS=80,443"
	@PORTSCAN_DEBUG=1 $(GOCMD) run $(MAIN_PATH) scan $(TARGET) --ports $(PORTS)

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

## release: Create a release build
release:
	@echo "Creating release build..."
	@goreleaser build --snapshot --clean
	@echo "✅ Release build completed"

## docker: Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker build -t $(BINARY_NAME):latest .
	@echo "✅ Docker image built: $(BINARY_NAME):$(VERSION)"

## docker-run: Run in Docker container
docker-run:
	@echo "Running in Docker..."
	@docker run --rm -it $(BINARY_NAME):latest scan $(TARGET) --ports $(PORTS)

## ci: Run full CI pipeline locally
ci: deps fmt lint test-race test-coverage security benchmark
	@echo "✅ Full CI pipeline completed successfully"

## quick: Quick development checks
quick: fmt lint test
	@echo "✅ Quick checks completed"

# Examples and usage
## examples: Show usage examples
examples:
	@echo "Usage Examples:"
	@echo ""
	@echo "  Development:"
	@echo "    make dev TARGET=localhost PORTS=80,443"
	@echo "    make dev TARGET=192.168.1.0/24 PORTS=1-1024"
	@echo ""
	@echo "  Testing:"
	@echo "    make test"
	@echo "    make benchmark"
	@echo "    make test-coverage"
	@echo ""
	@echo "  Building:"
	@echo "    make build"
	@echo "    make build-all"
	@echo "    make release"
	@echo ""
	@echo "  Quality:"
	@echo "    make lint"
	@echo "    make security"
	@echo "    make ci"

# Variables for development
TARGET ?= localhost
PORTS ?= 1-1024
