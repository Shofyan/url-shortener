package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ProcessingTime middleware adds X-Processing-Time-Micros header to all responses
func ProcessingTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		microseconds := time.Since(start).Microseconds()
		c.Header("X-Processing-Time-Micros", strconv.FormatInt(microseconds, 10))
	}
}