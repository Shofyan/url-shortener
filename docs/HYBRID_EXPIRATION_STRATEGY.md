# Hybrid Expiration Strategy for URL Shortener

## Executive Summary

This document describes the implementation of a **Hybrid Expiration Strategy** for the URL Shortener system that ensures correctness, scalability, and operational efficiency under read-heavy traffic.

The strategy combines **Lazy Validation** on the read path with **Asynchronous Background Cleanup** (The Reaper) to meet all functional and non-functional requirements.

## Architecture Overview

### Core Principles

1. **Read Path is Sacred**: Redirect requests must never perform synchronous delete operations
2. **Lazy Validation**: Expiration is checked logically on every read, regardless of physical data presence
3. **Background Cleanup**: Physical deletion happens asynchronously in batches
4. **Cache Defense**: Tombstones prevent thundering herd problems on hot expired URLs

### Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Request  │───▶│   Read Path     │───▶│   Cache/DB      │
│                 │    │ (Lazy Valid.)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               ▼ (No Sync Deletes!)
                       ┌─────────────────┐
                       │ Return HTTP 410 │
                       │   (If Expired)  │
                       └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│Background Timer │───▶│ Cleanup Service │───▶│ Batch Deletion  │
│  (15 minutes)   │    │  (The Reaper)   │    │  (1000/batch)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Implementation Details

### 1. Read Path: Lazy Validation

**File**: [`internal/application/usecase/shorten_url_usecase.go`](internal/application/usecase/shorten_url_usecase.go#L239)

**Key Changes**:
- Removed synchronous `urlRepo.Delete()` calls from `GetLongURL()`
- Added structured cache entries with expiration metadata
- Implemented tombstone caching for expired URLs
- Added logical expiration validation even for cached entries

**Flow**:
```go
// Phase 1: Check Cache
cacheEntry := cache.GetCacheEntry(key)
if cacheEntry.IsTombstone {
    return HTTP_410_Gone
}
if cacheEntry.IsExpired() {
    cache.SetTombstone(key, "expired", 1hour)
    return HTTP_410_Gone
}

// Phase 2: Check Database (cache miss)
url := db.FindByShortKey(key)
if url.IsExpired() {
    cache.SetTombstone(key, "expired", 1hour)
    // NO SYNC DELETE HERE!
    return HTTP_410_Gone
}
```

### 2. Background Cleanup: The Reaper

**Files**:
- [`internal/domain/service/cleanup_service.go`](internal/domain/service/cleanup_service.go)
- [`internal/domain/service/background_cleanup_service.go`](internal/domain/service/background_cleanup_service.go)

**Key Features**:
- Runs every 15 minutes (configurable)
- Processes 1000 URLs per batch (configurable)
- 1-hour buffer time prevents clock skew issues
- Batch deletion using SQL `DELETE WHERE short_key = ANY($1)`
- Graceful shutdown support

**Flow**:
```go
// Find expired URLs with buffer time
cutoffTime := time.Now().Add(-1 * time.Hour)
expiredURLs := repo.FindExpiredURLs(cutoffTime, 1000)

// Batch delete in single transaction
shortKeys := extractShortKeys(expiredURLs)
repo.DeleteExpiredBatch(shortKeys)

// Clean cache entries (best effort)
for key := range cacheKeys {
    cache.Delete(key)
}
```

### 3. Cache Strategy: Structured Entries + Tombstones

**Files**:
- [`internal/domain/repository/cache_repository.go`](internal/domain/repository/cache_repository.go)
- [`internal/infrastructure/cache/redis/cache_repository.go`](internal/infrastructure/cache/redis/cache_repository.go)

**Structured Cache Entry**:
```go
type CacheEntry struct {
    LongURL     string     `json:"long_url"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    IsTombstone bool       `json:"is_tombstone"`
    Reason      string     `json:"reason,omitempty"` // "expired", "deleted"
}
```

**Benefits**:
- **Payload Validation**: Expiration checked even on cache hits
- **Tombstone Protection**: Prevents thundering herd on hot expired URLs
- **Cache TTL Alignment**: Redis TTL matches URL expiration time
- **Clock Skew Safety**: Application logic validates regardless of cache TTL

### 4. Database Extensions

**Files**:
- [`internal/domain/repository/url_repository.go`](internal/domain/repository/url_repository.go)
- [`internal/infrastructure/database/postgres/url_repository.go`](internal/infrastructure/database/postgres/url_repository.go)

**New Methods**:
```go
// Efficient cleanup operations
FindExpiredURLs(ctx, before, maxResults) ([]*URL, error)
DeleteExpiredBatch(ctx, shortKeys) error
GetExpiredCount(ctx, before) (int64, error)
```

**SQL Optimizations**:
- B-Tree index on `expires_at` column
- Batch deletion using `ANY()` operator
- Transaction-wrapped batch operations

## Configuration

**File**: [`config.yaml`](config.yaml)

```yaml
app:
  # Hybrid expiration strategy settings
  cleanupenabled: true        # Enable/disable cleanup service
  cleanupinterval: "15m"      # How often to run cleanup
  cleanupbatchsize: 1000      # URLs processed per batch
  cleanupbuffertime: "1h"     # Safety buffer for clock skew
  cleanupmaxduration: "5m"    # Max time per cleanup operation
```

## Operational Benefits

### Latency Impact
- **Read Path**: ✅ Zero blocking operations, nanosecond timestamp checks
- **Write Path**: ✅ Background cleanup doesn't impact redirect traffic
- **Cache Defense**: ✅ Tombstones prevent database spikes on hot expired URLs

### Storage Control
- **Growth**: ✅ Controlled by background cleanup (max 15min*batch_size growth)
- **Cleanup**: ✅ Efficient batch deletion with proper indexing
- **Monitoring**: ✅ Cleanup stats available via `/api/admin/cleanup/stats`

### Consistency Guarantees
- **Correctness**: ✅ No expired URL is ever redirected (lazy validation)
- **Cache Safety**: ✅ Expiration validated on both cache hits and misses
- **Race Conditions**: ✅ Safe under concurrent access (no shared state on read path)

### Scalability Features
- **Horizontal**: ✅ Cleanup service can run on multiple instances safely
- **Load Balancing**: ✅ Read path has no cross-instance dependencies
- **Failure Resilience**: ✅ Cleanup failures don't affect redirect availability

## Testing & Validation

### Critical Test
**File**: [`tests/unit/usecase/hybrid_expiration_critical_test.go`](tests/unit/usecase/hybrid_expiration_critical_test.go)

**Test**: `TestHybridExpirationStrategy_NoSynchronousDeletes`

This test verifies the **core principle** that expired URLs return HTTP 410 without performing synchronous database deletes.

```bash
$ go test -run TestHybridExpirationStrategy_NoSynchronousDeletes ./tests/unit/usecase/hybrid_expiration_critical_test.go ./tests/unit/usecase/mocks.go -v
=== RUN   TestHybridExpirationStrategy_NoSynchronousDeletes
--- PASS: TestHybridExpirationStrategy_NoSynchronousDeletes (0.00s)
PASS
```

### Cleanup Service Tests
**File**: [`tests/unit/service/background_cleanup_service_test.go`](tests/unit/service/background_cleanup_service_test.go)

Tests batch cleanup, buffer time handling, statistics tracking, and configuration validation.

## Monitoring & Observability

### Cleanup Statistics
**Endpoint**: `GET /api/admin/cleanup/stats`

```json
{
  "last_cleanup_time": "2026-01-17T10:30:00Z",
  "total_cleaned": 15420,
  "last_batch_size": 847,
  "successful_runs": 234,
  "failed_runs": 2,
  "average_cleanup_ms": 156.7,
  "is_running": true
}
```

### Key Metrics to Monitor
- `total_cleaned`: Total URLs cleaned up over time
- `failed_runs`: Cleanup failures (should be near zero)
- `average_cleanup_ms`: Performance trending
- `last_batch_size`: Storage growth rate indicator

## Advanced Optimizations (Future)

### Database Partitioning
Instead of individual row deletes, partition tables by expiration date:
```sql
-- Partition by month
CREATE TABLE urls_2026_01 PARTITION OF urls
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- Cleanup becomes:
DROP TABLE urls_2025_12; -- Instant, reclaims disk space
```

### Hot URL Protection
For viral URLs that expire during peak traffic:
- Cache negative results with longer TTL
- Use Redis pub/sub for cache invalidation coordination
- Implement circuit breaker patterns for database protection

### Analytics Integration
- Log cleanup events to time-series database
- Track cleanup efficiency and storage trends
- Alert on abnormal expiration patterns

## Conclusion

The Hybrid Expiration Strategy successfully meets all requirements:

- ✅ **Functional**: Expired URLs return HTTP 410, active URLs redirect properly
- ✅ **Performance**: Read path remains low-latency and non-blocking
- ✅ **Scalability**: Supports horizontal scaling and high throughput
- ✅ **Operational**: Safe under concurrent access with comprehensive monitoring

The implementation provides a robust foundation for production-scale URL shortening with proper expiration handling that can be extended with additional features like analytics, archiving, or advanced caching strategies.
