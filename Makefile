.PHONY: help build run test clean docker-up docker-down migrate migrate-create migrate-status

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/url-shortener cmd/api/main.go

build-migrate: ## Build the migration tool
	@echo "Building migration tool..."
	@go build -o bin/migrate cmd/migrate/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs: ## View Docker logs
	@docker-compose logs -f

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go

migrate-docker: ## Run migrations inside Docker
	@echo "Running migrations in Docker..."
	@docker-compose exec app go run cmd/migrate/main.go

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at;"

db-shell: ## Open PostgreSQL shell
	@echo "Opening database shell..."
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -c "DROP DATABASE IF EXISTS $(POSTGRES_DB);"
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -c "CREATE DATABASE $(POSTGRES_DB);"
	@echo "Database reset complete. Run 'make migrate' to apply migrations."

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

.DEFAULT_GOAL := help
