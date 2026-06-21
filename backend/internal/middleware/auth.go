package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"matchlab/backend/internal/auth"
)

// RequireAuth validates a Bearer JWT and stores its identity in Gin context.
func RequireAuth(parser auth.TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortInvalidToken(c)
			return
		}
		claims, err := parser.Parse(parts[1])
		if err != nil {
			abortInvalidToken(c)
			return
		}
		c.Set(auth.ContextUserIDKey, claims.Subject)
		c.Set(auth.ContextRoleKey, claims.Role)
		c.Next()
	}
}

func abortInvalidToken(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":   "invalid_token",
		"message": "登录凭证无效或已过期",
	})
}
