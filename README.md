# URL Shortener

A high-performance URL shortener service built with Go, following Clean Architecture and Domain-Driven Design (DDD) principles.

## ⚡ Quick Start

Get the application running in 5 minutes:

```bash
# 1. Clone the repository
git clone https://github.com/Shofyan/url-shortener.git
cd url-shortener

# 2. Copy environment file
cp .env.example .env

# 3. Start Docker containers (PostgreSQL + Redis)
make docker-up

# 4. Run database migrations
make migrate

# 5. Start the application
make run

# 6. Test it!
curl http://localhost:8080/health
```

🎉 The API is now running at `http://localhost:8080`

## Features

- 🚀 Shorten long URLs to compact, shareable links
- 🔄 Automatic redirection from short to long URLs
- 🎯 Custom short key support (alphanumeric, hyphens, underscores)
- ⏰ Optional URL expiration
- 📊 Visit count tracking and statistics
- 💾 PostgreSQL for persistent storage
- ⚡ Redis caching for fast redirects
- 🔒 Rate limiting to prevent abuse
- 📝 Comprehensive logging for debugging and monitoring
- 🏗️ Clean Architecture with DDD principles
- 🐳 Docker support with docker-compose

## Architecture

### Clean Architecture Layers

```
cmd/
  api/                    # Application entry point
internal/
  domain/                 # Domain Layer (Business Logic)
    entity/              # Domain entities
    valueobject/         # Value objects
    repository/          # Repository interfaces
    service/             # Domain services
  application/            # Application Layer (Use Cases)
    dto/                 # Data Transfer Objects
    usecase/             # Application use cases
  infrastructure/         # Infrastructure Layer (External Dependencies)
    config/              # Configuration management
    database/            # Database implementations
    cache/               # Cache implementations
    generator/           # ID generation implementations
  interfaces/             # Interface Layer (API/HTTP)
    http/
      handler/           # HTTP handlers
      middleware/        # HTTP middleware
      router/            # Route configuration
```

### Key Design Patterns

- **Domain-Driven Design (DDD)**: Core business logic in domain layer
- **Repository Pattern**: Abstract data access
- **Dependency Injection**: Loose coupling between layers
- **Value Objects**: Immutable domain primitives
- **Use Cases**: Application-specific business rules

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Cache**: Redis
- **ID Generation**: Snowflake + Base62 encoding
- **Configuration**: Viper

## Getting Started

### Prerequisites

**Option 1: Using Docker (Recommended)**
- Docker Desktop or Docker Engine
- Docker Compose
- Make (optional, for convenience commands)

**Option 2: Manual Setup**
- Go 1.21 or higher
- PostgreSQL 14+
- Redis 6+

### Quick Start with Docker (Recommended)

Follow these steps to run the complete application stack using Docker:

#### Step 1: Clone the Repository

```bash
git clone https://github.com/Shofyan/url-shortener.git
cd url-shortener
```

#### Step 2: Configure Environment Variables

Create a `.env` file from the example:

```bash
# Copy the example file
cp .env.example .env

# Edit if needed (default values work for local development)
```

The `.env` file contains all necessary configuration:
- Database credentials (PostgreSQL)
- Redis settings
- Server configuration
- Docker container names

#### Step 3: Start Docker Containers

Start PostgreSQL and Redis containers:

```bash
# Using Make (recommended)
make docker-up

# Or using docker-compose directly
docker-compose up -d
```

This will start:
- ✅ PostgreSQL database on port 5432
- ✅ Redis cache on port 6379

Wait a few seconds for containers to be healthy, then verify:

```bash
# Check container status
docker ps

# Should see 2 running containers:
# - url-shortener-postgres
# - url-shortener-redis
```

#### Step 4: Run Database Migrations

Apply database schema migrations:

```bash
# Using Make (recommended)
make migrate

# Or run directly with Go
go run cmd/migrate/main.go
```

Expected output:
```
Connecting to database...
✓ Database connection successful
Found 2 pending migration(s)
✓ Applied migration: 001_create_urls_table.sql
✓ Applied migration: 002_create_analytics_table.sql
✓ All migrations completed successfully
```

Verify migrations were applied:

```bash
make migrate-status
```

#### Step 5: Run the Application

Start the URL shortener service:

```bash
# Using Make
make run

# Or run directly with Go
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

#### Step 6: Test the API

```bash
# Health check
curl http://localhost:8080/health

# Create a short URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/very/long/url"}'
```

### Full Docker Stack (Optional)

To run everything including the app in Docker:

```bash
# Build and start all services (postgres, redis, app)
docker-compose up -d

# View logs
make docker-logs

# Stop all services
make docker-down
```

### Stopping the Application

```bash
# Stop Docker containers
make docker-down

# Or to remove all data volumes
docker-compose down -v
```

### Manual Installation (Without Docker)

1. Clone the repository:
```bash
git clone https://github.com/Shofyan/url-shortener.git
cd url-shortener
```

2. Install dependencies:
```bash
go mod download
```

3. Set up PostgreSQL database:
```bash
createdb urlshortener
psql urlshortener < internal/infrastructure/database/migrations/001_create_urls_table.sql
```

4. Start Redis:
```bash
redis-server
```

5. Configure the application:
Edit `config.yaml` with your database and Redis settings.

6. Run the application:
```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

## Database Migrations

### Overview

The application uses a custom migration system to manage database schema changes. Migrations are versioned SQL files that are applied in order.

### Migration Files Location

```
internal/infrastructure/database/migrations/
├── 001_create_urls_table.sql          # Initial schema
├── 002_create_analytics_table.sql     # Analytics tracking
└── 003_alter_short_key_length.sql     # Increased short_key to VARCHAR(12)
```

### Running Migrations

#### Using Make Commands (Recommended)

```bash
# Run all pending migrations
make migrate

# Check migration status
make migrate-status

# Output shows applied migrations:
#           version           |         applied_at
# ----------------------------+----------------------------
#  001_create_urls_table      | 2025-12-29 10:18:29.053216
#  002_create_analytics_table | 2025-12-29 10:18:29.06039
```

#### Using Go Directly

```bash
# Run all pending migrations
go run cmd/migrate/main.go
```

#### Database Management Commands

```bash
# Open PostgreSQL shell in Docker
make db-shell

# Reset database (drops and recreates)
make db-reset

# View database tables
docker exec url-shortener-postgres psql -U postgres -d urlshortener -c "\dt"
```

### Migration Features

- ✅ **Automatic tracking**: Uses `schema_migrations` table to track applied migrations
- ✅ **Idempotent**: Safe to run multiple times (only applies pending migrations)
- ✅ **Transactional**: Each migration runs in a transaction (rolls back on error)
- ✅ **Versioned**: Migrations are applied in order based on filename
- ✅ **Environment-aware**: Reads configuration from `.env` file

### Creating New Migrations

Create a new SQL file in the migrations directory:

```bash
# File naming convention: ###_description.sql
# Example:
internal/infrastructure/database/migrations/003_add_user_accounts.sql
```

Example migration file:

```sql
-- Add new column to urls table
ALTER TABLE urls ADD COLUMN user_id BIGINT;

-- Create index
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
```

After creating the file, run migrations:

```bash
make migrate
```

### Environment Variables

The migration tool uses these environment variables from `.env`:

```bash
DATABASE_HOST       # Default: localhost
DATABASE_PORT       # Default: 5432
DATABASE_USER       # Default: postgres
DATABASE_PASSWORD   # Default: postgres
DATABASE_DBNAME     # Default: urlshortener
POSTGRES_USER       # For Docker containers
POSTGRES_PASSWORD   # For Docker containers
POSTGRES_DB         # For Docker containers
```

## Make Commands Reference

The project includes a Makefile with convenient commands for development. Run `make help` to see all available commands.

### Application Commands

```bash
make build              # Build the application binary
make run                # Run the application locally
make clean              # Clean build artifacts
```

### Docker Commands

```bash
make docker-up          # Start Docker containers (PostgreSQL + Redis)
make docker-down        # Stop Docker containers
make docker-logs        # View Docker container logs
```

### Database Commands

```bash
make migrate            # Run database migrations
make migrate-status     # Check which migrations have been applied
make migrate-docker     # Run migrations inside Docker container
make db-shell           # Open PostgreSQL shell (requires Docker)
make db-reset           # Reset database (drop and recreate)
```

### Testing Commands

```bash
make test               # Run all tests
make test-coverage      # Run tests with coverage report
make lint               # Run linter
make fmt                # Format code
```

### Dependencies

```bash
make deps               # Download and tidy Go dependencies
```

### Example Workflow

```bash
# 1. Start infrastructure
make docker-up

# 2. Run migrations
make migrate

# 3. Check migration status
make migrate-status

# 4. Run the application
make run

# 5. In another terminal, run tests
make test

# 6. Stop infrastructure when done
make docker-down
```

## API Endpoints

### Testing the API

Once the application is running (via Docker Compose or manually), you can test the endpoints:

```bash
# Health check
curl http://localhost:8080/health
```

### Create Short URL

**Create a short URL with auto-generated key:**

```bash
POST /api/shorten
Content-Type: application/json

{
  "long_url": "https://example.com/very/long/url"
}
```

Response:
```json
{
  "short_url": "http://localhost:8080/2O994sNdbYu",
  "short_key": "2O994sNdbYu",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2025-12-29T10:00:00Z"
}
```

**Create a short URL with custom key:**

```bash
POST /api/shorten
Content-Type: application/json

{
  "long_url": "https://github.com/Shofyan/url-shortener",
  "custom_key": "my-github"
}
```

Response:
```json
{
  "short_url": "http://localhost:8080/my-github",
  "short_key": "my-github",
  "long_url": "https://github.com/Shofyan/url-shortener",
  "created_at": "2025-12-29T10:00:00Z"
}
```

**Custom Key Requirements:**
- ✅ Alphanumeric characters (a-z, A-Z, 0-9)
- ✅ Hyphens (-) and underscores (_)
- ✅ Maximum 12 characters
- ❌ Special characters (!, @, #, $, etc.)
- ❌ Spaces

**Examples of Valid Custom Keys:**
- `my-link`
- `custom_key_123`
- `MyCustomURL`
- `test-2024`
- `user_profile`

**Create a short URL with expiration:**

```bash
POST /api/shorten
Content-Type: application/json

{
  "long_url": "https://example.com/very/long/url",
  "custom_key": "temp-link",
  "expires_in": 86400  // 24 hours in seconds
}
```

Response:
```json
{
  "short_url": "http://localhost:8080/temp-link",
  "short_key": "temp-link",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2025-12-29T10:00:00Z",
  "expires_at": "2025-12-30T10:00:00Z"
}
```

**Important Notes:**
- Without a custom key, duplicate long URLs return the existing short URL
- With a custom key, a new short URL is always created (allowing multiple short URLs for the same long URL)
- Auto-generated keys use Snowflake IDs encoded in Base62 (11 characters)

### Redirect Short URL

```bash
GET /:shortKey
```

Returns a 302 redirect to the original long URL.

### Get URL Statistics

```bash
GET /api/stats/:shortKey
```

Response:
```json
{
  "short_key": "abc123",
  "long_url": "https://example.com/very/long/url",
  "visit_count": 42,
  "created_at": "2025-12-29T10:00:00Z",
  "expires_at": "2025-12-30T10:00:00Z"
}
```

### Health Check

```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "url-shortener"
}
```

## Configuration

Configuration can be set via `config.yaml` or environment variables:

```yaml
server:
  port: "8080"
  readtimeout: "10s"
  writetimeout: "10s"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "urlshortener"

redis:
  host: "localhost"
  port: "6379"

app:
  baseurl:Compose Usage

### Starting the Application

```bash
# Start all services in detached mode
docker-compose up -d

# Start and view logs
docker-compose up

# Start and rebuild containers
docker-compose up --build
```

### Viewing Logs

```bash
# View all logs
docker-compose logs

# View logs for specific service
docker-compose logs app
docker-compose logs postgres
docker-compose logs redis

# Follow logs in real-time
docker-compose logs -f app
```

### Managing Services

```bash
# Stop all services (keeps data)
docker-compose stop

# Start stopped services
docker-compose start

# Restart all services
docker-compose restart

# Stop and remove containers (keeps volumes)
docker-compose down

# Remove everything including volumes (⚠️ deletes all data)
docker-compose down -v
```

### Service Health Checks

```bash
# Check service status
docker-compose ps

# Execute commands in running container
docker-compose exec app /bin/sh

# Check PostgreSQL connection
docker-compose exec postgres psql -U postgres -d urlshortener -c "SELECT COUNT(*) FROM urls;"

# Check Redis connection
docker-compose exec redis redis-cli ping
```

### Environment Variables

You can override configuration using environment variables in a `.env` file:

```bash
# .env file
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=your_secure_password
DATABASE_DBNAME=urlshortener

REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

APP_BASEURL=http://localhost:8080
```

### Docker Standalone (Without Compose)

If you prefer to run without Docker Compose:

```bash
# Build the image
docker build -t url-shortener .

# Run PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=urlshortener \
  -p 5432:5432 \
  postgres:15-alpine

# Run Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:7-alpine

# Run the application
docker run -d --name url-shortener \
  -p 8080:8080 \
  -e DATABASE_HOST=host.docker.internal \
  -e REDIS_HOST=host.docker.internal \
  url-shortener
### ID Generation: Snowflake + Base62

The URL shortener uses a sophisticated ID generation system:

- **Snowflake Algorithm**: Generates unique 64-bit IDs that are:
  - Distributed and conflict-free across multiple instances
  - Time-sortable (IDs increase with time)
  - High-performance (thousands per second)
  
- **Base62 Encoding**: Converts Snowflake IDs into short, URL-safe strings
  - Character set: `0-9A-Za-z` (62 characters)
- **Error handling**: Continues operation even if cache fails

### Logging & Monitoring

The application includes comprehensive logging throughout the request lifecycle:

- **Request/Response logging**: All HTTP requests with status, latency, and IP
- **Error tracking**: Detailed error messages with context
- **Operation tracing**: Step-by-step logging in critical paths (URL shortening, retrieval)
- **Performance metrics**: Execution time and bottleneck identification

**Log Format:**
```
2025/12/29 03:47:39 [Shorten] Starting URL shortening process for: https://example.com
2025/12/29 03:47:39 [Shorten] Normalized URL: https://example.com
2025/12/29 03:47:39 [Shorten] Generated short key: 2O994sNdbYu, ID: 2005485841350135808
2025/12/29 03:47:39 [Shorten] URL saved successfully to database
2025/12/29 03:47:39 [POST] /api/shorten | Status: 200 | Latency: 5.602086ms | IP: 172.18.0.1
```
  - Typical length: 11 characters
  - Example: `2005485841350135808` → `2O994sNdbYu`
  
- **Benefits**: 
  - No database lookups needed for ID generation
  - Zero collision risk
  - Scalable across multiple servers
  - Deterministic (same ID always produces same short key)

### Short Key Format

- **Auto-generated**: 11 characters using Base62 encoding of Snowflake IDs
- **Custom keys**: 1-12 characters, alphanumeric plus hyphens and underscores
- **Validation**: Ensures keys are URL-safe and meet length requirements
- **Uniqueness**: Custom keys are checked for conflicts before creation

### Caching Strategy

- **Cache-Aside Pattern**: Check cache first, fallback to DB
- **TTL-based expiration**: Respects URL expiration times
- **Write-through**: Cache on creation for immediate availability

### Rate Limiting

- **Per-IP rate limiting** using token bucket algorithm
- Prevents abuse and ensures fair usage
- Configurable limits per endpoint

## Performance Considerations

- **Database Indexing**: Optimized indexes on short_key and long_url
- **Connection Pooling**: Efficient DB and Redis connection management
- **Async Operations**: Non-blocking visit count updates
- **Horizontal Scalability**: Stateless design for easy scaling

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/domain/valueobject
```

### API Testing with Postman

A complete Postman collection is provided for testing all API endpoints:

**Files:**
- `URL-Shortener.postman_collection.json` - Complete API collection
- `URL-Shortener-Local.postman_environment.json` - Local environment variables
- `URL-Shortener-Production.postman_environment.json` - Production environment template

**Import to Postman:**

1. Open Postman
2. Click **Import** button
3. Select the collection file: `URL-Shortener.postman_collection.json`
4. Import environment: `URL-Shortener-Local.postman_environment.json`
5. Select the "URL Shortener - Local" environment in Postman

**Collection Features:**
- ✅ All API endpoints with examples
- ✅ Automated tests for each request
- ✅ Environment variables for easy switching between local/production
- ✅ Example responses (success and error cases)
- ✅ Request descriptions and documentation

**Included Endpoints:**
- Health Check
- Shorten URL (basic, custom key, with expiration)
- Redirect to Long URL
- Get URL Statistics
- Error handling examples

**Using the Collection:**

```bash
# 1. Start the application
docker-compose up -d

# 2. Run the "Health Check" request to verify the service is running
# 3. Run "Shorten URL" to create a short link (saves short_key to environment)
# 4. Run "Get URL Statistics" to see visit count
# 5. Run "Redirect to Long URL" to test the redirect
```

The collection automatically saves the `short_key` from successful shorten requests, making it easy to test the redirect and stats endpoints.

## Docker Support

```bash
# Build image
docker build -t url-shortener .

# Run with docker-compose
docker-compose up
```

## Trade-offs & Future Improvements

### Current Trade-offs

- **301 vs 302 Redirects**: Using 302 (temporary) for tracking; 301 is faster but cached by browsers
- **Eventual Consistency**: Cache and DB may briefly differ
- **Visit Count Accuracy**: Async updates may lose counts in crashes

### Future Enhancements

- [ ] Analytics dashboard
- [ ] URL preview before redirect
- [ ] Bulk URL shortening API
- [ ] QR code generation
- [ ] User accounts and URL management
- [ ] A/B testing support
- [ ] Geographical analytics

## Troubleshooting

### Common Issues and Solutions

#### 1. Migration Status Failed

**Error**: `make migrate-status` returns error "The system cannot find the file specified"

**Solution**: Make sure Docker containers are running:
```bash
make docker-up
docker ps  # Should show url-shortener-postgres container
```

#### 2. Connection Refused to Database

**Error**: `Failed to connect to database: connection refused`

**Solution**: 
```bash
# Check if PostgreSQL container is running
docker ps | grep postgres

# If not running, start containers
make docker-up

# Wait 10-15 seconds for PostgreSQL to be ready
# Then run migrations
make migrate
```

#### 3. Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find what's using the port
# Windows PowerShell:
netstat -ano | findstr :8080
netstat -ano | findstr :5432

# Kill the process or change port in .env file
# Edit .env and change SERVER_PORT or DATABASE_PORT
```

#### 4. Migrations Not Applied


#### 7. Short Key Length Errors

**Error**: `pq: value too long for type character varying(10)`

**Solution**: This means your database schema needs updating:
```bash
# Apply the migration to increase short_key length
docker compose exec postgres psql -U postgres -d urlshortener -c "ALTER TABLE urls ALTER COLUMN short_key TYPE VARCHAR(12);"

# Or rebuild with updated migrations
docker compose down -v
docker compose up -d
```

#### 8. Invalid Short Key Format

**Error**: `invalid short key format`

**Solution**: Custom keys must follow these rules:
- Only alphanumeric characters (a-z, A-Z, 0-9)
- Can include hyphens (-) and underscores (_)
- Maximum 12 characters
- No spaces or special characters

**Valid examples**: `my-link`, `custom_key`, `Test123`
**Invalid examples**: `my link`, `key@123`, `special!char`
**Error**: Migration files exist but `make migrate-status` shows nothing

**Solution**:
```bash
# Ensure you're connected to the right database
make migrate

# If migrations fail, reset database
make db-reset
make migrate
```

#### 5. Docker Compose Errors

**Error**: `Error response from daemon: conflict`

**Solution**:
```bash
# Stop and remove all containers
make docker-down

# Remove volumes if needed (WARNING: deletes all data)
docker-compose down -v

# Start fresh
make docker-up
```

#### 6. Cannot Access Database from Host

**Problem**: `make migrate` works but application can't connect

**Solution**: Ensure `.env` has correct settings:
```bash
# For running app locally with Docker DB
DATABASE_HOST=localhost
DATABASE_PORT=5432

# For running app inside Docker
DATABASE_HOST=postgres  # service name in docker-compose.yml
```

### Getting Help

If you encounter issues:

1. **Check logs**: `make docker-logs` to see container output
2. **Check Docker status**: `docker ps` to verify containers are running
3. **Verify environment**: Ensure `.env` file exists and has correct values
4. **Database connection**: Run `make db-shell` to test direct database access
5. **Clean slate**: Run `make docker-down`, `docker-compose down -v`, then start fresh

### Useful Debug Commands

```bash
# Check all environment variables loaded
make help  # Shows all available commands

# View PostgreSQL logs
docker logs url-shortener-postgres

# View Redis logs
docker logs url-shortener-redis

# Connect to database and check tables
make db-shell
\dt              # List all tables
\d urls          # Describe urls table
SELECT * FROM schema_migrations;  # Check applied migrations
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
