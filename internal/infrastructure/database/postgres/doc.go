// Package postgres provides PostgreSQL database implementation for the URL shortener service.
//
// This package implements the repository interfaces using PostgreSQL as the
// persistent storage backend. It handles database connections, query execution,
// and transaction management while providing efficient data access patterns
// for URL operations.
//
// Features:
//   - Connection pooling and management
//   - Prepared statements for optimal performance
//   - Transaction support for data consistency
//   - Error handling and connection recovery
//   - Query optimization and indexing strategies
//   - Migration support and schema management
//
// Database operations supported:
//   - URL creation with unique constraint handling
//   - URL retrieval by short key with caching integration
//   - Analytics data collection and aggregation
//   - Batch operations for performance optimization
//   - Data expiration and cleanup procedures
//
// The PostgreSQL implementation provides ACID compliance and
// scalable data storage for the URL shortener service.
package postgres
