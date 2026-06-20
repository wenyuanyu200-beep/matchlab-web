// Package router assembles the public HTTP routes.
package router

import (
	"github.com/gin-gonic/gin"

	"matchlab/backend/internal/health"
)

// New returns the application HTTP handler.
func New() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	api := engine.Group("/api")
	api.GET("/health", health.Handler)

	return engine
}
