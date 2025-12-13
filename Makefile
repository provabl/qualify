.PHONY: help check test integration build build-all clean install-tools docker-build docker-up docker-down web-dev web-build web-test web-test-unit web-test-e2e web-coverage

# Variables
VERSION ?= dev
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commitSHA=$(COMMIT_SHA) -X main.buildDate=$(BUILD_DATE)

# Binary names
AGENT_BIN := bin/ark-agent
BACKEND_BIN := bin/ark-backend
CLI_BIN := bin/ark

# Go settings
GOFLAGS := -trimpath
TESTFLAGS := -race -coverprofile=coverage.out -covermode=atomic

## help: Show this help message
help:
	@echo "Ark - AWS Research Kit"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development:"
	@echo "  check           Fast pre-commit checks (fmt, vet, lint, short tests)"
	@echo "  test            Full unit tests with coverage"
	@echo "  integration     Integration tests (requires LocalStack)"
	@echo "  install-tools   Install development tools"
	@echo ""
	@echo "Building:"
	@echo "  build           Build all binaries (agent, backend, CLI)"
	@echo "  build-agent     Build agent only"
	@echo "  build-backend   Build backend only"
	@echo "  build-cli       Build CLI only"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build    Build Docker images"
	@echo "  docker-up       Start development environment"
	@echo "  docker-down     Stop development environment"
	@echo ""
	@echo "Web:"
	@echo "  web-dev         Start web development server"
	@echo "  web-build       Build web production bundle"
	@echo "  web-test        Run all web tests (unit + E2E)"
	@echo "  web-test-unit   Run web unit tests only"
	@echo "  web-test-e2e    Run web E2E tests only"
	@echo "  web-coverage    Generate web test coverage report"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean           Remove build artifacts"

## check: Fast pre-commit checks
check:
	@echo "→ Running gofmt..."
	@gofmt -l -w .
	@echo "→ Running go vet..."
	@go vet ./...
	@echo "→ Running staticcheck..."
	@staticcheck ./...
	@echo "→ Running short tests..."
	@go test -short -race ./...
	@echo "✓ All checks passed"

## test: Full unit tests with coverage
test:
	@echo "→ Running full test suite..."
	@go test $(TESTFLAGS) ./...
	@echo "→ Coverage report:"
	@go tool cover -func=coverage.out | tail -1

## integration: Integration tests
integration:
	@echo "→ Running integration tests (requires LocalStack)..."
	@go test -race -tags=integration -timeout=10m ./...

## build: Build all binaries
build: build-agent build-backend build-cli
	@echo "✓ All binaries built"

## build-agent: Build agent binary
build-agent:
	@echo "→ Building agent..."
	@mkdir -p bin
	@go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(AGENT_BIN) ./cmd/ark-agent
	@echo "✓ Agent built: $(AGENT_BIN)"

## build-backend: Build backend binary
build-backend:
	@echo "→ Building backend..."
	@mkdir -p bin
	@go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BACKEND_BIN) ./cmd/ark-backend
	@echo "✓ Backend built: $(BACKEND_BIN)"

## build-cli: Build CLI binary
build-cli:
	@echo "→ Building CLI..."
	@mkdir -p bin
	@go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(CLI_BIN) ./cmd/ark
	@echo "✓ CLI built: $(CLI_BIN)"

## install-tools: Install development tools
install-tools:
	@echo "→ Installing development tools..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Tools installed"

## docker-build: Build Docker images
docker-build:
	@echo "→ Building Docker images..."
	@docker build -t ark-agent:$(VERSION) -f docker/Dockerfile.agent .
	@docker build -t ark-backend:$(VERSION) -f docker/Dockerfile.backend .
	@echo "✓ Docker images built"

## docker-up: Start development environment (LocalStack + services)
docker-up:
	@echo "→ Starting development environment..."
	@docker-compose -f docker/docker-compose.dev.yml up -d
	@echo "✓ Development environment running"
	@echo ""
	@echo "Services:"
	@echo "  LocalStack:  http://localhost:4566"
	@echo "  Backend:     http://localhost:8080"
	@echo "  Agent:       http://localhost:8737"
	@echo "  Web:         http://localhost:5173"

## docker-down: Stop development environment
docker-down:
	@echo "→ Stopping development environment..."
	@docker-compose -f docker/docker-compose.dev.yml down
	@echo "✓ Development environment stopped"

## web-dev: Start web development server
web-dev:
	@echo "→ Starting web development server..."
	@cd web && npm run dev

## web-build: Build web production bundle
web-build:
	@echo "→ Building web production bundle..."
	@cd web && npm run build
	@echo "✓ Web bundle built: web/dist/"

## web-test: Run all web tests (unit + E2E)
web-test: web-test-unit web-test-e2e
	@echo "✓ All web tests passed"

## web-test-unit: Run web unit tests
web-test-unit:
	@echo "→ Running web unit tests..."
	@cd web && npm run test:unit:run
	@echo "✓ Web unit tests passed"

## web-test-e2e: Run web E2E tests
web-test-e2e:
	@echo "→ Running web E2E tests..."
	@cd web && npm run test:e2e
	@echo "✓ Web E2E tests passed"

## web-coverage: Generate web test coverage report
web-coverage:
	@echo "→ Generating web test coverage..."
	@cd web && npm run test:unit:coverage
	@echo "✓ Coverage report generated in web/coverage/"

## clean: Remove build artifacts
clean:
	@echo "→ Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out
	@rm -rf web/dist/
	@rm -rf web/coverage/
	@rm -rf web/playwright-report/
	@rm -rf web/test-results/
	@echo "✓ Cleaned"

# Development workflow targets
.PHONY: dev agent-dev backend-dev cli-dev

## dev: Start full development environment (all services)
dev: docker-up
	@echo ""
	@echo "Development environment ready!"
	@echo "Run 'make agent-dev' in one terminal"
	@echo "Run 'make backend-dev' in another terminal"
	@echo "Run 'make web-dev' in a third terminal"

## agent-dev: Run agent in development mode (auto-reload)
agent-dev:
	@echo "→ Running agent in development mode..."
	@go run ./cmd/ark-agent

## backend-dev: Run backend in development mode (auto-reload)
backend-dev:
	@echo "→ Running backend in development mode..."
	@go run ./cmd/ark-backend

## cli-dev: Run CLI in development mode
cli-dev:
	@echo "→ Running CLI in development mode..."
	@go run ./cmd/ark $(ARGS)

# Release targets
.PHONY: release-build release-test

## release-build: Build release binaries for all platforms
release-build:
	@echo "→ Building release binaries for version $(VERSION)..."
	@mkdir -p dist
	# Linux amd64
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-agent-linux-amd64 ./cmd/ark-agent
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-backend-linux-amd64 ./cmd/ark-backend
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-linux-amd64 ./cmd/ark
	# Linux arm64
	@GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-agent-linux-arm64 ./cmd/ark-agent
	@GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-backend-linux-arm64 ./cmd/ark-backend
	@GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-linux-arm64 ./cmd/ark
	# macOS amd64
	@GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-agent-darwin-amd64 ./cmd/ark-agent
	@GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-darwin-amd64 ./cmd/ark
	# macOS arm64
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-agent-darwin-arm64 ./cmd/ark-agent
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-darwin-arm64 ./cmd/ark
	# Windows amd64
	@GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-agent-windows-amd64.exe ./cmd/ark-agent
	@GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o dist/ark-windows-amd64.exe ./cmd/ark
	@echo "✓ Release binaries built in dist/"

## release-test: Run all tests before release
release-test: check test integration web-test
	@echo "✓ All release tests passed"
