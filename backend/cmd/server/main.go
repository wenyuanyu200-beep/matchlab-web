package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"matchlab/backend/internal/config"
	"matchlab/backend/internal/database"
	"matchlab/backend/internal/router"
	"matchlab/backend/internal/user"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)
	if cfg.UsesDevelopmentJWTSecret() {
		log.Print("WARNING: JWT_SECRET is using the development fallback; set a strong secret in production")
	}

	userRepository := user.NewGormRepository(nil)
	if cfg.Database.Configured() {
		db, err := database.Open(cfg.Database)
		if err != nil {
			log.Printf("database unavailable; continuing without it: %v", err)
		} else {
			userRepository = user.NewGormRepository(db)
			sqlDB, err := db.DB()
			if err != nil {
				log.Printf("database pool unavailable: %v", err)
			} else {
				defer func() {
					if err := sqlDB.Close(); err != nil {
						log.Printf("close database: %v", err)
					}
				}()
				log.Printf("connected to PostgreSQL (%s)", database.SafeDescription(cfg.Database))
			}
		}
	} else {
		log.Print("database configuration incomplete; starting HTTP server without PostgreSQL")
	}

	server := &http.Server{
		Addr: cfg.Address(),
		Handler: router.New(router.Dependencies{
			Users:     userRepository,
			JWTSecret: cfg.JWTSecret,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("MatchLab API listening on http://%s", cfg.Address())
		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signals:
		log.Printf("received %s; shutting down", sig)
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server failed: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
}
