// Package router assembles the public HTTP routes.
package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/auth"
	"matchlab/backend/internal/health"
	"matchlab/backend/internal/middleware"
	"matchlab/backend/internal/user"
)

// Dependencies contains external services needed by HTTP routes.
type Dependencies struct {
	DB         *gorm.DB
	Users      user.Repository
	Activities activity.Repository
	JWTSecret  string
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
	activities := dependencies.Activities
	if activities == nil {
		activities = activity.NewGormRepository(dependencies.DB)
	}
	tokens := auth.NewTokenManager(dependencies.JWTSecret)
	authHandler := auth.NewHandler(auth.NewService(users, tokens))
	activityHandler := activity.NewHandler(activities)

	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	api.GET("/me", middleware.RequireAuth(tokens), authHandler.Me)

	api.GET("/activities", activityHandler.List)
	api.GET("/activities/:id", activityHandler.Detail)
	api.POST("/activities", middleware.RequireAuth(tokens), activityHandler.Create)
	api.GET("/me/activities", middleware.RequireAuth(tokens), activityHandler.MyActivities)
	api.POST("/activities/:id/apply", middleware.RequireAuth(tokens), activityHandler.Apply)
	api.GET("/me/applications", middleware.RequireAuth(tokens), activityHandler.MyApplications)
	api.GET("/activities/:id/applications", middleware.RequireAuth(tokens), activityHandler.ActivityApplications)
	api.POST("/applications/:id/approve", middleware.RequireAuth(tokens), activityHandler.Approve)
	api.POST("/applications/:id/reject", middleware.RequireAuth(tokens), activityHandler.Reject)

	return engine
}
