#!/bin/bash

# Database Migration Script for URL Shortener

set -e

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}URL Shortener - Database Migration${NC}"
echo "======================================"

# Check if running in Docker
if [ -f /.dockerenv ]; then
    echo "Running inside Docker container"
    DB_HOST="${DATABASE_HOST:-postgres}"
    DB_PORT="${DATABASE_PORT:-5432}"
    DB_USER="${DATABASE_USER:-postgres}"
    DB_PASS="${DATABASE_PASSWORD:-postgres}"
    DB_NAME="${DATABASE_DBNAME:-urlshortener}"
else
    echo "Running on host machine"
    DB_HOST="${DATABASE_HOST:-localhost}"
    DB_PORT="${DATABASE_PORT:-5432}"
    DB_USER="${DATABASE_USER:-postgres}"
    DB_PASS="${DATABASE_PASSWORD:-postgres}"
    DB_NAME="${DATABASE_DBNAME:-urlshortener}"
fi

export DATABASE_DSN="host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASS dbname=$DB_NAME sslmode=disable"

echo ""
echo "Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# Run migrations
echo -e "${YELLOW}Running migrations...${NC}"
go run cmd/migrate/main.go

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Migration completed successfully!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Migration failed!${NC}"
    exit 1
fi
