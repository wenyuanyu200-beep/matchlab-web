// Package health provides service liveness endpoints.
package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler reports whether the HTTP process is running. It intentionally does
// not check PostgreSQL so load balancers can distinguish process health from
// dependency readiness.
func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "MatchLab API running",
	})
}
