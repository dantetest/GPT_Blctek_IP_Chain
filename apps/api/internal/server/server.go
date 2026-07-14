package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/health"
	appmiddleware "github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)
func New(logger *slog.Logger, db *sql.DB, cache *redis.Client) http.Handler {
	router := gin.New(); router.Use(gin.Recovery(), appmiddleware.RequestID(), accessLog(logger))
	h := health.NewHandler(db, cache); api := router.Group("/api/v1"); api.GET("/health/live", h.Live); api.GET("/health", h.Ready); return router
}
func accessLog(logger *slog.Logger) gin.HandlerFunc { return func(c *gin.Context) { c.Next(); requestID,_:=c.Get("request_id"); logger.Info("http request", "method",c.Request.Method,"path",c.Request.URL.Path,"status",c.Writer.Status(),"request_id",requestID) } }
