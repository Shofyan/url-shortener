package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter middleware implements rate limiting per IP.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter middleware.
func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(float64(requestsPerMinute) / 60.0),
		burst:    burst,
	}

	// Cleanup old visitors every 3 minutes
	go rl.cleanupVisitors()

	return rl
}

// Limit returns the rate limiting middleware.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()

		v, exists := rl.visitors[ip]
		if !exists {
			limiter := rate.NewLimiter(rl.rate, rl.burst)
			rl.visitors[ip] = &visitor{limiter, time.Now()}
			v = rl.visitors[ip]
		}

		v.lastSeen = time.Now()
		rl.mu.Unlock()

		if !v.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()

			return
		}

		c.Next()
	}
}

// cleanupVisitors removes old visitors from the map.
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
