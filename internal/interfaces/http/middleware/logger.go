package middleware

import (
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

// Logger middleware logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[ERROR] %v", e.Err)
			}
		}

		log.Printf("[%s] %s %s | Status: %d | Latency: %v",
			c.Request.Method,
			path,
			query,
			c.Writer.Status(),
			latency,
		)
	}
}
