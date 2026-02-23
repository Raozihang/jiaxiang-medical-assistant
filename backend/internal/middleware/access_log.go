package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		requestID := GetRequestID(c)

		log.Printf(
			"request_id=%s method=%s path=%s status=%d latency_ms=%d",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency.Milliseconds(),
		)
	}
}
