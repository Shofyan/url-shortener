// Package url_shortener provides a high-performance URL shortening service built with Go.
//
// This service allows users to shorten long URLs into compact, shareable links and provides
// analytics for tracking click statistics. The service is built using clean architecture
// principles with clear separation between domain logic, application use cases, and infrastructure.
//
// Key Features:
//   - URL shortening with customizable short codes
//   - Click analytics and statistics
//   - Redis caching for high performance
//   - PostgreSQL for persistent storage
//   - Rate limiting and middleware support
//   - REST API with JSON responses
//   - Web interface for easy URL shortening
//
// Architecture:
//   - Domain Layer: Core business logic and entities
//   - Application Layer: Use cases and DTOs
//   - Infrastructure Layer: Database, cache, and external services
//   - Interface Layer: HTTP handlers, middleware, and routing
//
// The service supports multiple ID generation strategies including Base62 encoding
// and Snowflake IDs for distributed environments.
//
// For more information, visit: https://github.com/Shofyan/url-shortener
package urlshortener
