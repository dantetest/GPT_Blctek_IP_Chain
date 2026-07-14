package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/auth"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/config"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/health"
	appmiddleware "github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func New(cfg config.Config, logger *slog.Logger, db *sql.DB, cache *redis.Client) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery(), appmiddleware.RequestID(), accessLog(logger))

	healthHandler := health.NewHandler(db, cache)
	tokenManager := auth.NewTokenManager(cfg.JWTIssuer, cfg.JWTSecret, cfg.AccessTokenTTL)
	authRepository := auth.NewRepository(db)
	authService := auth.NewService(authRepository, tokenManager, auth.ServiceConfig{
		RefreshTokenTTL:  cfg.RefreshTokenTTL,
		AutoVerifyEmail:  cfg.AutoVerifyEmail,
		LoginMaxAttempts: cfg.LoginMaxAttempts,
		LoginLockTTL:     cfg.LoginLockTTL,
	})
	authHandler := auth.NewHandler(authService)

	api := router.Group("/api/v1")
	api.GET("/health/live", healthHandler.Live)
	api.GET("/health", healthHandler.Ready)

	authRoutes := api.Group("/auth")
	authRoutes.POST("/register", appmiddleware.RateLimit(cache, "auth-register", 5, time.Minute, appmiddleware.ClientIP), authHandler.Register)
	authRoutes.POST("/login", appmiddleware.RateLimit(cache, "auth-login", 10, time.Minute, appmiddleware.ClientIP), authHandler.Login)
	authRoutes.POST("/refresh", appmiddleware.RateLimit(cache, "auth-refresh", 30, time.Minute, appmiddleware.ClientIP), authHandler.Refresh)
	authRoutes.POST("/logout", authHandler.Logout)
	authRoutes.POST("/verify-email", appmiddleware.RateLimit(cache, "auth-verify", 10, time.Minute, appmiddleware.ClientIP), authHandler.VerifyEmail)

	protected := api.Group("")
	protected.Use(appmiddleware.Authenticate(tokenManager))
	protected.GET("/me", authHandler.Me)

	return router
}

func accessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		requestID, _ := c.Get("request_id")
		logger.Info(
			"http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"request_id", requestID,
		)
	}
}
