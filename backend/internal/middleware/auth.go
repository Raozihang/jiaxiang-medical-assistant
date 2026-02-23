package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiaxiang-medical-assistant/backend/internal/service"
)

func AuthRequired(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authService == nil {
			writeAuthError(c)
			return
		}

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

		claims, err := authService.VerifyToken(token)
		if err != nil {
			writeAuthError(c)
			return
		}

		c.Set("auth_subject", claims.Subject)
		c.Set("auth_role", claims.Role)
		c.Set("auth_name", claims.Name)
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
