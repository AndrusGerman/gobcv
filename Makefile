# BCV Currency API Makefile

# Variables
BINARY_NAME=api
BINARY_DIR=bin
MAIN_PATH=cmd/api/main.go
BINARY_PATH=${BINARY_DIR}/${BINARY_NAME}

# Go parameters
GOCMD=go
GOBUILD=${GOCMD} build
GOCLEAN=${GOCMD} clean
GOTEST=${GOCMD} test
GOGET=${GOCMD} get
GOMOD=${GOCMD} mod

# Build the binary
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BINARY_DIR}
	${GOBUILD} -o ${BINARY_PATH} ${MAIN_PATH}
	@echo "✅ Build completed: ${BINARY_PATH}"

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building ${BINARY_NAME} for Windows..."
	@mkdir -p ${BINARY_DIR}
	GOOS=windows GOARCH=amd64 ${GOBUILD} -o ${BINARY_PATH}.exe ${MAIN_PATH}
	@echo "✅ Windows build completed: ${BINARY_PATH}.exe"

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building ${BINARY_NAME} for Linux..."
	@mkdir -p ${BINARY_DIR}
	GOOS=linux GOARCH=amd64 ${GOBUILD} -o ${BINARY_PATH}-linux ${MAIN_PATH}
	@echo "✅ Linux build completed: ${BINARY_PATH}-linux"

# Build for macOS
.PHONY: build-darwin
build-darwin:
	@echo "Building ${BINARY_NAME} for macOS..."
	@mkdir -p ${BINARY_DIR}
	GOOS=darwin GOARCH=amd64 ${GOBUILD} -o ${BINARY_PATH}-darwin ${MAIN_PATH}
	@echo "✅ macOS build completed: ${BINARY_PATH}-darwin"

# Build for all platforms
.PHONY: build-all
build-all: build-windows build-linux build-darwin
	@echo "✅ All platform builds completed"

# Run the application
.PHONY: run
run: build
	@echo "Starting ${BINARY_NAME}..."
	./${BINARY_PATH}

# Run without building (development)
.PHONY: dev
dev:
	@echo "Starting in development mode..."
	${GOCMD} run ${MAIN_PATH}

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	${GOCLEAN}
	rm -rf ${BINARY_DIR}
	@echo "✅ Clean completed"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	${GOMOD} download
	${GOMOD} tidy
	@echo "✅ Dependencies installed"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	${GOCMD} fmt ./...
	@echo "✅ Code formatted"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	${GOTEST} -v ./...
	@echo "✅ Tests completed"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	${GOTEST} -v -coverprofile=coverage.out ./...
	${GOCMD} tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Lint code
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "✅ Linting completed"

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	gosec ./...
	@echo "✅ Security check completed"

# API health check
.PHONY: health
health:
	@echo "Checking API health..."
	@curl -s http://localhost:8080/api/v1/health | jq '.' || echo "❌ API not responding"

# Test all endpoints
.PHONY: test-api
test-api:
	@echo "Testing API endpoints..."
	@echo "1. Health check:"
	@curl -s http://localhost:8080/api/v1/health | jq '.'
	@echo "\n2. Get all currencies:"
	@curl -s http://localhost:8080/api/v1/currencies | jq '.'
	@echo "\n3. Refresh currencies:"
	@curl -s -X POST http://localhost:8080/api/v1/currencies/refresh | jq '.'
	@echo "✅ API tests completed"

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t bcv-currency-api .
	@echo "✅ Docker image built"

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 bcv-currency-api

# Generate project documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "📚 Documentation server started at http://localhost:6060"

# Setup development environment
.PHONY: setup
setup: deps fmt
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✅ Development environment ready"

# Show project structure
.PHONY: tree
tree:
	@echo "Project structure:"
	@tree -I 'bin|*.exe|go.sum|.git' --dirsfirst

# Show help
.PHONY: help
help:
	@echo "BCV Currency API - Available commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  build         Build the binary for current platform"
	@echo "  build-windows Build for Windows"
	@echo "  build-linux   Build for Linux"
	@echo "  build-darwin  Build for macOS"
	@echo "  build-all     Build for all platforms"
	@echo ""
	@echo "Development commands:"
	@echo "  run           Build and run the application"
	@echo "  dev           Run in development mode (no build)"
	@echo "  deps          Install dependencies"
	@echo "  fmt           Format code"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "Quality commands:"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  lint          Run linter"
	@echo "  security      Run security checks"
	@echo ""
	@echo "API commands:"
	@echo "  health        Check API health"
	@echo "  test-api      Test all API endpoints"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run Docker container"
	@echo ""
	@echo "Documentation:"
	@echo "  docs          Generate and serve documentation"
	@echo "  tree          Show project structure"
	@echo "  help          Show this help message"
	@echo ""
	@echo "Setup:"
	@echo "  setup         Setup development environment"

# Default target
.DEFAULT_GOAL := help
