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

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		normalized := strings.TrimSpace(role)
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		roleValue, ok := c.Get("auth_role")
		if !ok {
			writeAuthError(c)
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			writeForbiddenError(c)
			return
		}
		if _, ok := allowed[role]; !ok {
			writeForbiddenError(c)
			return
		}

		c.Next()
	}
}

func writeAuthError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":       1002,
		"message":    "未授权访问",
		"data":       gin.H{},
		"request_id": GetRequestID(c),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func writeForbiddenError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"code":       1003,
		"message":    "权限不足",
		"data":       gin.H{},
		"request_id": GetRequestID(c),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}
