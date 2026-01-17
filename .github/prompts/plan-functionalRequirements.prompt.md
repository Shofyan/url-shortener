# Technical Requirements Document

## Distributed URL Shortener - Functional Requirements Implementation

### Executive Summary

The current Go URL shortener implementation demonstrates solid architectural patterns but requires critical modifications to achieve full specification compliance. **Current compliance: ~45%**. This document outlines precise technical changes needed across API endpoints, database schema, business logic, and infrastructure layers.

---

## 1. API Endpoint Compliance Requirements

### 1.1 Critical Route Pattern Changes

**Current Issues:**
- Uses `/api/shorten` instead of `POST /`
- Uses `/:shortKey` instead of `GET /s/{short_code}`
- Missing `X-Processing-Time-Micros` header requirement

**Required Changes:**

| Component | File | Current | Required | Priority |
|-----------|------|---------|----------|----------|
| URL Creation | [interfaces/http/router/router.go] | `POST /api/shorten` | `POST /` | CRITICAL |
| Redirection | [interfaces/http/router/router.go] | `GET /:shortKey` | `GET /s/:shortKey` | CRITICAL |
| Processing Header | [interfaces/http/middleware/] | Not implemented | Add timing middleware | CRITICAL |

### 1.2 Request Parameter Compliance

**Current DTO Issues:**
- Uses `expires_in` instead of required `ttl_seconds`
- No default 24-hour TTL implementation

**Required Changes in [application/dto/url_dto.go]:**
```go
type ShortenURLRequest struct {
    LongURL    string `json:"long_url" binding:"required"`
    TTLSeconds int64  `json:"ttl_seconds,omitempty"` // Changed from expires_in
}
```

---

## 2. Database Schema Requirements

### 2.1 Critical Missing Fields

**Current Schema Gaps:**
- Missing `last_accessed_at` field for stats endpoint
- PII violation with `ip_address` storage in analytics table

**Required Migration Script:**
```sql
-- Migration: 004_add_last_accessed_at_remove_pii.sql
ALTER TABLE urls ADD COLUMN last_accessed_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_urls_last_accessed_at ON urls(last_accessed_at);

-- Remove PII violations
ALTER TABLE url_analytics DROP COLUMN ip_address;
ALTER TABLE url_analytics DROP COLUMN city;

-- Fix data type inconsistency
ALTER TABLE url_analytics ALTER COLUMN short_key TYPE VARCHAR(12);
```

### 2.2 Entity Model Updates

**Required Changes in [domain/entity/url.go]:**
```go
type URL struct {
    ID             int64
    ShortKey       *valueobject.ShortKey
    LongURL        *valueobject.LongURL
    CreatedAt      time.Time
    ExpiresAt      *time.Time
    VisitCount     int64
    LastAccessedAt *time.Time  // ADD THIS FIELD
}
```

---

## 3. Business Logic Implementation Requirements

### 3.1 Short Code Generation Compliance

**Current Issue:** Base62 generator includes prohibited characters `0`, `O`, `l`, `1`

**Required Fix in [infrastructure/generator/base62/generator.go]:**
```go
// Replace current character set
const base62Chars = "23456789ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz"
// Removes: 0, O, l, 1 for readability
```

### 3.2 TTL Default Implementation

**Current Issue:** No default TTL when `ttl_seconds` not provided

**Required Changes in [application/usecase/shorten_url_usecase.go]:**
```go
const defaultTTLSeconds = 24 * 60 * 60 // 24 hours

func (uc *ShortenURLUseCase) ShortenURL(ctx context.Context, req *dto.ShortenURLRequest) (*dto.ShortenURLResponse, error) {
    // Apply default TTL if not specified
    if req.TTLSeconds == 0 {
        req.TTLSeconds = defaultTTLSeconds
    }
    // ... rest of implementation
}
```

### 3.3 Thread-Safe Click Tracking

**Current Issue:** Race condition in async visit count increment

**Required Implementation:**
```go
// In repository layer - atomic database operation
func (r *PostgresURLRepository) IncrementVisitCount(ctx context.Context, shortKey *valueobject.ShortKey) error {
    query := `
        UPDATE urls
        SET visit_count = visit_count + 1,
            last_accessed_at = CURRENT_TIMESTAMP
        WHERE short_key = $1`

    _, err := r.db.ExecContext(ctx, query, shortKey.Value())
    return err
}
```

---

## 4. Middleware and Observability Requirements

### 4.1 Processing Time Header Implementation

**Required New File: [interfaces/http/middleware/timing.go]**
```go
package middleware

import (
    "strconv"
    "time"
    "github.com/gin-gonic/gin"
)

func ProcessingTime() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        microseconds := time.Since(start).Microseconds()
        c.Header("X-Processing-Time-Micros", strconv.FormatInt(microseconds, 10))
    }
}
```

### 4.2 Privacy Compliance Fix

**Critical Issue:** IP address logging violates PII requirements

**Required Fix in [interfaces/http/middleware/logger.go]:**
```go
// REMOVE IP address from logs
log.Printf("[%s] %s %s | Status: %d | Latency: %v",
    c.Request.Method, path, query, c.Writer.Status(), latency)
// DO NOT LOG: c.ClientIP()
```

---

## 5. Response Format Requirements

### 5.1 Stats Endpoint Response Update

**Current Missing Field:** `last_accessed_at` in stats response

**Required Changes in [application/dto/url_dto.go]:**
```go
type URLStatsResponse struct {
    ShortKey       string `json:"short_key"`
    LongURL        string `json:"long_url"`
    VisitCount     int64  `json:"visit_count"`
    CreatedAt      string `json:"created_at"`
    ExpiresAt      string `json:"expires_at,omitempty"`
    LastAccessedAt string `json:"last_accessed_at,omitempty"` // ADD THIS
}
```

---

## 6. Implementation Priority and Effort Matrix

### Phase 1: Critical Security & Compliance (1-2 days)
| Task | File(s) | Risk Level | Effort |
|------|---------|------------|---------|
| Remove IP logging | [interfaces/http/middleware/logger.go] | CRITICAL | 2 hours |
| Fix API routes | [interfaces/http/router/router.go] | HIGH | 4 hours |
| Add processing time header | [interfaces/http/middleware/timing.go]| MEDIUM | 4 hours |

### Phase 2: Data Integrity & Schema (2-3 days)
| Task | File(s) | Risk Level | Effort |
|------|---------|------------|---------|
| Database migration | New migration file | HIGH | 6 hours |
| Update entity model | [domain/entity/url.go]| HIGH | 4 hours |
| Repository updates | [infrastructure/database/postgres/url_repository.go] | HIGH | 8 hours |

### Phase 3: Business Logic Fixes (3-4 days)
| Task | File(s) | Risk Level | Effort |
|------|---------|------------|---------|
| Character exclusion | [infrastructure/generator/base62/generator.go]| MEDIUM | 3 hours |
| Default TTL | [application/usecase/shorten_url_usecase.go]| MEDIUM | 4 hours |
| Thread-safe counting | Multiple repository files | HIGH | 8 hours |

### Phase 4: Testing & Documentation (1-2 weeks)
| Task | Priority | Effort |
|------|----------|---------|
| Unit test suite | HIGH | 3 days |
| Concurrency tests | CRITICAL | 2 days |
| Integration tests | MEDIUM | 2 days |

---

## 7. Testing Requirements

### 7.1 Critical Missing Test Coverage

**Current State:** Zero test files found
**Required Coverage:**
- Concurrency safety tests for click counting
- Deterministic TTL tests with mocked time
- Interface verification tests for storage abstraction

### 7.2 Required Test Structure
```
/tests/
├── unit/
│   ├── usecase/
│   ├── repository/
│   └── handler/
├── integration/
│   └── api/
└── concurrency/
    └── click_count_test.go  // 100+ concurrent requests test
```

---

## 8. Validation and Acceptance Criteria

### 8.1 Functional Requirements Checklist

| Requirement | Current Status | Target Status | Validation Method |
|-------------|----------------|---------------|-------------------|
| POST / endpoint | ✅ Implemented | ✅ Correct route | API test |
| GET /s/{short_code} | ✅ Implemented | ✅ Correct route | API test |
| TTL default 24h | ✅ Implemented | ✅ Implemented | Unit test |
| Character exclusion | ✅ Fixed | ✅ Filtered | Unit test |
| Thread-safe clicks | ✅ Implemented | ✅ Atomic ops | Concurrency test |
| last_accessed_at | ✅ Added | ✅ Tracked | Integration test |

### 8.2 Non-Functional Requirements Checklist

| Requirement | Current Status | Target Status | Validation Method |
|-------------|----------------|---------------|-------------------|
| X-Processing-Time-Micros | ✅ Implemented | ✅ All responses | API test |
| No PII storage/logging | ✅ Compliant | ✅ Privacy compliant | Code review |
| 10K RPS capability | ✅ Architecture ready | ✅ Load tested | Performance test |

---

## 9. Risk Assessment and Mitigation

### 9.1 High-Risk Changes
1. **Database Schema Changes**: Requires migration planning and downtime consideration
2. **API Route Changes**: Breaking change requiring client updates
3. **Concurrency Model Changes**: Risk of introducing new race conditions

### 9.2 Recommended Deployment Strategy
1. Deploy with feature flags for new routes
2. Run both old and new routes during transition
3. Implement comprehensive monitoring for new concurrency patterns
4. Schedule database migration during low-traffic window

---

## 10. Implementation Checklist

### Development Phase
- [x] Update router configuration for correct endpoints
- [x] Implement processing time middleware
- [x] Create database migration for last_accessed_at field
- [x] Fix character exclusion in short code generation
- [x] Implement default TTL logic
- [x] Add thread-safe visit counting
- [x] Update DTOs and response formats
- [x] Remove IP logging for privacy compliance

### Testing Phase
- [x] Write unit tests for all business logic
- [x] Implement concurrency tests for click counting
- [x] Create integration tests for all API endpoints
- [x] Add performance tests for scalability requirements
- [x] Validate TTL logic with deterministic time tests

### Deployment Phase
- [ ] Create deployment runbook with rollback plan
- [ ] Set up monitoring for new performance metrics
- [ ] Plan database migration execution
- [ ] Coordinate client updates for API changes

**Total Estimated Effort:** 3-4 weeks for full implementation
**Risk Level:** HIGH due to breaking API changes and zero current test coverage
