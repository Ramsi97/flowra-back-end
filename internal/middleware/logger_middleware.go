package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs structured information about each request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// Process request
		c.Next()

		// Access latency
		latency := time.Since(t)

		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Log structured message
		log.Printf("[GIN] %s | %d | %13v | %s | %s",
			method,
			status,
			latency,
			path,
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)

		// Log any errors that were added to the context
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[ERROR] %v", e.Err)
			}
		}
	}
}
