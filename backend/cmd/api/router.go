package main

import (
	"context"
	"net/http"
	"time"

	"basepro/backend/internal/auth"
	"basepro/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type AppDeps struct {
	DB          *sqlx.DB
	Version     string
	AuthHandler *auth.Handler
	JWTManager  *auth.JWTManager
}

func newRouter(deps AppDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		statusCode := http.StatusOK
		payload := gin.H{
			"status":  "ok",
			"version": deps.Version,
			"db":      "up",
		}

		healthCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if deps.DB == nil || deps.DB.PingContext(healthCtx) != nil {
			statusCode = http.StatusServiceUnavailable
			payload["status"] = "degraded"
			payload["db"] = "down"
		}

		c.JSON(statusCode, payload)
	})

	api.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": deps.Version})
	})

	if deps.AuthHandler != nil && deps.JWTManager != nil {
		authGroup := api.Group("/auth")
		authGroup.POST("/login", deps.AuthHandler.Login)
		authGroup.POST("/refresh", deps.AuthHandler.Refresh)
		authGroup.POST("/logout", deps.AuthHandler.Logout)
		authGroup.GET("/me", middleware.JWTAuth(deps.JWTManager), deps.AuthHandler.Me)
	}

	return r
}
