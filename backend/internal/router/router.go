// Package router assembles the public HTTP routes.
package router

import (
	"github.com/gin-gonic/gin"

	"matchlab/backend/internal/auth"
	"matchlab/backend/internal/health"
	"matchlab/backend/internal/middleware"
	"matchlab/backend/internal/user"
)

// Dependencies contains external services needed by HTTP routes.
type Dependencies struct {
	Users     user.Repository
	JWTSecret string
}

// New returns the application HTTP handler.
func New(dependencies Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	api := engine.Group("/api")
	api.GET("/health", health.Handler)

	users := dependencies.Users
	if users == nil {
		users = user.NewGormRepository(nil)
	}
	tokens := auth.NewTokenManager(dependencies.JWTSecret)
	authHandler := auth.NewHandler(auth.NewService(users, tokens))

	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	api.GET("/me", middleware.RequireAuth(tokens), authHandler.Me)

	return engine
}
