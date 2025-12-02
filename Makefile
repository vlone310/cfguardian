.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down sqlc-generate lint format install-tools

# Variables
APP_NAME=cfguardian
MAIN_PATH=./cmd/server
BINARY_PATH=./bin/$(APP_NAME)
DOCKER_IMAGE=cfguardian:latest
DB_URL?=postgres://postgres:postgres@localhost:5432/cfguardian?sslmode=disable

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

## help: Show this help message
help:
	@echo '$(GREEN)GoConfig Guardian - Development Commands$(NC)'
	@echo ''
	@echo 'Usage:'
	@echo '  $(YELLOW)make$(NC) $(GREEN)<target>$(NC)'
	@echo ''
	@echo 'Targets:'
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/##//g' | awk 'BEGIN {FS = ":"}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

## install-tools: Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/air-verse/air@latest
	@echo "$(GREEN)Tools installed successfully!$(NC)"

## setup-hooks: Setup Git hooks
setup-hooks:
	@echo "$(GREEN)Setting up Git hooks...$(NC)"
	./scripts/setup-hooks.sh
	@echo "$(GREEN)Git hooks installed!$(NC)"

## build: Build the application
build:
	@echo "$(GREEN)Building application...$(NC)"
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BINARY_PATH)$(NC)"

## run: Run the application
run:
	@echo "$(GREEN)Running application...$(NC)"
	go run $(MAIN_PATH)/main.go

## dev: Run with live reload (requires air)
dev:
	@echo "$(GREEN)Starting development server with live reload...$(NC)"
	air

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests complete!$(NC)"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "$(GREEN)Generating coverage report...$(NC)"
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

## test-unit: Run unit tests only
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	go test -v -race -short ./...

## test-integration: Run integration tests
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	go test -v -race -run Integration ./test/integration/...

## test-e2e: Run end-to-end tests
test-e2e:
	@echo "$(GREEN)Running E2E tests...$(NC)"
	go test -v -race ./test/e2e/...

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

## lint: Run linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	golangci-lint run ./...

## format: Format code
format:
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	goimports -w .

## tidy: Tidy go modules
tidy:
	@echo "$(GREEN)Tidying go modules...$(NC)"
	go mod tidy
	go mod verify

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete!$(NC)"

## sqlc-generate: Generate code from SQL
sqlc-generate:
	@echo "$(GREEN)Generating Go code from SQL...$(NC)"
	sqlc generate

## migrate-create: Create a new migration (usage: make migrate-create NAME=create_users)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=create_users$(NC)"; \
		exit 1; \
	fi
	@./scripts/migrate.sh create $(NAME)

## migrate-up: Run all pending database migrations
migrate-up:
	@./scripts/migrate.sh up

## migrate-up-one: Run one pending migration
migrate-up-one:
	@./scripts/migrate.sh up 1

## migrate-down: Rollback last migration
migrate-down:
	@./scripts/migrate.sh down 1

## migrate-down-all: Rollback ALL migrations (DANGEROUS)
migrate-down-all:
	@./scripts/migrate.sh down

## migrate-force: Force migration version (usage: make migrate-force VERSION=1)
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required. Usage: make migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	@./scripts/migrate.sh force $(VERSION)

## migrate-version: Show current migration version
migrate-version:
	@./scripts/migrate.sh version

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE) -f docker/Dockerfile .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE)$(NC)"

## docker-up: Start Docker Compose services
docker-up:
	@echo "$(GREEN)Starting Docker services...$(NC)"
	docker-compose -f docker/docker-compose.yml up -d
	@echo "$(GREEN)Services started!$(NC)"

## docker-down: Stop Docker Compose services
docker-down:
	@echo "$(YELLOW)Stopping Docker services...$(NC)"
	docker-compose -f docker/docker-compose.yml down
	@echo "$(GREEN)Services stopped!$(NC)"

## docker-logs: View Docker Compose logs
docker-logs:
	docker-compose -f docker/docker-compose.yml logs -f

## docker-clean: Remove Docker containers and volumes
docker-clean:
	@echo "$(YELLOW)Cleaning Docker resources...$(NC)"
	docker-compose -f docker/docker-compose.yml down -v
	@echo "$(GREEN)Docker cleanup complete!$(NC)"

## api-generate: Generate API code from OpenAPI spec
api-generate:
	@echo "$(GREEN)Generating API code from OpenAPI spec...$(NC)"
	oapi-codegen -package generated -generate types,chi-server,strict-server api/openapi.yaml > api/generated/api.gen.go
	@echo "$(GREEN)API code generated!$(NC)"

## docs-serve: Serve API documentation
docs-serve:
	@echo "$(GREEN)Serving API documentation at http://localhost:8080$(NC)"
	@which redoc-cli > /dev/null || npm install -g redoc-cli
	redoc-cli serve api/openapi.yaml --watch

## k8s-apply: Apply Kubernetes manifests
k8s-apply:
	@echo "$(GREEN)Applying Kubernetes manifests...$(NC)"
	kubectl apply -f k8s/

## k8s-delete: Delete Kubernetes resources
k8s-delete:
	@echo "$(YELLOW)Deleting Kubernetes resources...$(NC)"
	kubectl delete -f k8s/

## setup: Initial setup for development
setup: install-tools setup-hooks
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo "$(GREEN)Setup complete! Edit .env with your configuration.$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Edit .env with your configuration"
	@echo "  2. Start services: make docker-up"
	@echo "  3. Run migrations: make migrate-up"
	@echo "  4. Start development: make dev"

## verify: Run all verification checks
verify: format lint test
	@echo "$(GREEN)All checks passed!$(NC)"

## all: Build and run the application
all: build run

.DEFAULT_GOAL := help

