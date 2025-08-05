# Go-Dig Build Configuration
# Makefile for building Windows executable

# Variables
BINARY_NAME=go-dig.exe
MAIN_PACKAGE=.
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

# Default target
.PHONY: all
all: clean build

# Build Windows executable
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for current platform (development)
.PHONY: build-dev
build-dev:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Development build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build release (optimized)
.PHONY: release
release: clean test
	@echo "Building release version..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Binary size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build (default)"
	@echo "  build        - Build Windows executable"
	@echo "  build-dev    - Build for current platform"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code (requires golangci-lint)"
	@echo "  deps         - Install/update dependencies"
	@echo "  release      - Build optimized release version"
	@echo "  help         - Show this help"