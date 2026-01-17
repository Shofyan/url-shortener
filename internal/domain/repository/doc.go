// Package repository defines interfaces for data persistence in the URL shortener domain.
//
// This package contains repository contracts that abstract data storage
// and retrieval operations. These interfaces define the persistence layer
// without coupling to specific database implementations, allowing for
// flexible storage strategies and easy testing through mocking.
//
// Repository interfaces include:
//   - URLRepository for URL entity persistence
//   - CacheRepository for high-performance data caching
//   - Query methods for data retrieval and filtering
//   - Transactional operations for data consistency
//
// Key principles:
//   - Database-agnostic interface definitions
//   - Clean separation from domain logic
//   - Support for different storage backends
//   - Optimized for both read and write operations
//   - Error handling and connection management
//
// These interfaces are implemented by concrete repositories in the
// infrastructure layer, maintaining clean architecture boundaries.
package repository
