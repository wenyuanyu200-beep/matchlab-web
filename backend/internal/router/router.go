// Package router assembles the public HTTP routes.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/admin"
	"matchlab/backend/internal/auth"
	"matchlab/backend/internal/health"
	matching "matchlab/backend/internal/match"
	"matchlab/backend/internal/middleware"
	"matchlab/backend/internal/questionnaire"
	"matchlab/backend/internal/user"
)

// Dependencies contains external services needed by HTTP routes.
type Dependencies struct {
	DB                 *gorm.DB
	Users              user.Repository
	Activities         activity.Repository
	Questionnaires     questionnaire.Repository
	Matches            matching.Repository
	Admin              admin.Repository
	JWTSecret          string
	CORSAllowedOrigins []string
}

// New returns the application HTTP handler.
func New(dependencies Dependencies) *gin.Engine {
	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	engine.Use(middleware.CORS(dependencies.CORSAllowedOrigins), gin.Logger(), gin.Recovery())
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "route not found"})
	})
	engine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method_not_allowed", "message": "method not allowed"})
	})

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
	questionnaires := dependencies.Questionnaires
	if questionnaires == nil {
		questionnaires = questionnaire.NewGormRepository(dependencies.DB)
	}
	matches := dependencies.Matches
	if matches == nil {
		matches = matching.NewGormRepository(dependencies.DB)
	}
	adminRepository := dependencies.Admin
	if adminRepository == nil {
		adminRepository = admin.NewGormRepository(dependencies.DB)
	}
	tokens := auth.NewTokenManager(dependencies.JWTSecret)
	authHandler := auth.NewHandler(auth.NewService(users, tokens))
	activityHandler := activity.NewHandler(activities)
	questionnaireHandler := questionnaire.NewHandlerWithService(questionnaire.NewService(questionnaires))
	matchHandler := matching.NewHandlerWithService(matching.NewService(matches))
	adminHandler := admin.NewHandler(adminRepository)
	authenticated := middleware.RequireAuth(tokens)

	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	api.GET("/me", authenticated, authHandler.Me)

	api.GET("/activities", activityHandler.List)
	api.GET("/activities/:id", activityHandler.Detail)
	api.POST("/activities", authenticated, activityHandler.Create)
	api.GET("/me/activities", authenticated, activityHandler.MyActivities)
	api.POST("/activities/:id/apply", authenticated, activityHandler.Apply)
	api.GET("/me/applications", authenticated, activityHandler.MyApplications)
	api.GET("/activities/:id/applications", authenticated, activityHandler.ActivityApplications)
	api.POST("/applications/:id/approve", authenticated, activityHandler.Approve)
	api.POST("/applications/:id/reject", authenticated, activityHandler.Reject)

	api.POST("/questionnaires", authenticated, questionnaireHandler.Submit)
	api.GET("/me/profile", authenticated, questionnaireHandler.Profile)
	api.POST("/match/recommend", authenticated, matchHandler.Recommend)
	api.GET("/me/matches", authenticated, matchHandler.MyMatches)

	adminRoutes := api.Group("/admin", authenticated, middleware.RequireAdmin())
	adminRoutes.GET("/stats", adminHandler.Stats)
	adminRoutes.GET("/users", adminHandler.Users)
	adminRoutes.GET("/activities", adminHandler.Activities)
	adminRoutes.GET("/applications", adminHandler.Applications)
	adminRoutes.GET("/feedbacks", adminHandler.Feedbacks)
	adminRoutes.POST("/users/:id/role", adminHandler.UpdateUserRole)

	return engine
}
