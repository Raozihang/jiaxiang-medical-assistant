package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			writeAuthError(c)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			writeAuthError(c)
			return
		}

		c.Next()
	}
}

func writeAuthError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":       1002,
		"message":    "unauthorized",
		"data":       gin.H{},
		"request_id": GetRequestID(c),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}
