# Database Migration Script for URL Shortener (PowerShell)

$ErrorActionPreference = "Stop"

Write-Host "URL Shortener - Database Migration" -ForegroundColor Yellow
Write-Host "======================================" -ForegroundColor Yellow
Write-Host ""

# Get database configuration from environment or use defaults
$DB_HOST = if ($env:DATABASE_HOST) { $env:DATABASE_HOST } else { "localhost" }
$DB_PORT = if ($env:DATABASE_PORT) { $env:DATABASE_PORT } else { "5432" }
$DB_USER = if ($env:DATABASE_USER) { $env:DATABASE_USER } else { "postgres" }
$DB_PASS = if ($env:DATABASE_PASSWORD) { $env:DATABASE_PASSWORD } else { "postgres" }
$DB_NAME = if ($env:DATABASE_DBNAME) { $env:DATABASE_DBNAME } else { "urlshortener" }

$env:DATABASE_DSN = "host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASS dbname=$DB_NAME sslmode=disable"

Write-Host "Database Configuration:"
Write-Host "  Host: $DB_HOST"
Write-Host "  Port: $DB_PORT"
Write-Host "  User: $DB_USER"
Write-Host "  Database: $DB_NAME"
Write-Host ""

# Run migrations
Write-Host "Running migrations..." -ForegroundColor Yellow
try {
    go run cmd/migrate/main.go
    Write-Host ""
    Write-Host "✓ Migration completed successfully!" -ForegroundColor Green
    exit 0
}
catch {
    Write-Host ""
    Write-Host "✗ Migration failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}
