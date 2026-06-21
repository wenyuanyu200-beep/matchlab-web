package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"matchlab/backend/internal/auth"
)

// RequireAdmin allows only authenticated contexts carrying the admin role.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(auth.ContextRoleKey) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "administrator access required",
			})
			return
		}
		c.Next()
	}
}
