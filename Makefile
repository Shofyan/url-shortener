.PHONY: help build run test clean docker-up docker-down migrate migrate-create migrate-status lint lint-fix lint-install pre-commit

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

lint: ## Run golangci-lint with gocyclo and revive enabled
	@echo "Running golangci-lint..."
	@golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --fix
lint-install: ## Install golangci-lint and pre-commit tools
	@echo "Installing linting tools..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		echo "golangci-lint already installed"; \
	else \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
	fi
	@if command -v pre-commit > /dev/null 2>&1; then \
		echo "pre-commit already installed"; \
		pre-commit install; \
	else \
		echo "❌ pre-commit not found"; \
		echo "➡ Install it using:"; \
		echo "   pip install pre-commit"; \
		echo "   # or"; \
		echo "   pipx install pre-commit"; \
		echo "   # or"; \
		echo "   brew install pre-commit"; \
		echo "Then run 'make lint-install' again"; \
	fi

pre-commit: ## Run pre-commit hooks on all files
	@echo "Running pre-commit hooks..."
	@pre-commit run --all-files

pre-commit-update: ## Update pre-commit hooks
	@echo "Updating pre-commit hooks..."
	@pre-commit autoupdate

quality: ## Run all quality checks (lint + pre-commit)
	@echo "Running all quality checks..."
	@$(MAKE) lint
	@$(MAKE) pre-commit

quality-fix: ## Run all quality fixes (lint-fix + fmt)
	@echo "Running all quality fixes..."
	@$(MAKE) lint-fix
	@$(MAKE) fmt

security-scan: ## Run security scan with gosec
	@echo "Running security scan..."
	@gosec ./...

complexity-check: ## Check cyclomatic complexity
	@echo "Checking cyclomatic complexity..."
	@gocyclo -over 10 .

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports > /dev/null 2>&1; then \
		echo "Running goimports..."; \
		goimports -w .; \
	fi

.DEFAULT_GOAL := help
