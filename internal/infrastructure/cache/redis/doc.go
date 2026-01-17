// Package redis provides Redis-based caching implementation for the URL shortener service.
//
// This package implements the cache repository interface using Redis as the
// backend storage system. It provides high-performance caching capabilities
// for frequently accessed URLs, reducing database load and improving response times.
//
// Features:
//   - Redis client connection management
//   - Automatic key expiration and TTL management
//   - Connection pooling for optimal performance
//   - Error handling and retry mechanisms
//   - Serialization and deserialization of cached data
//   - Cache invalidation strategies
//
// The Redis cache implementation supports:
//   - URL lookup caching with configurable expiration
//   - Hot data preloading for better performance
//   - Memory-efficient storage with compression
//   - Cluster support for high availability
//
// This package abstracts Redis-specific details while providing
// a clean interface that matches the domain's caching requirements.
package redis
