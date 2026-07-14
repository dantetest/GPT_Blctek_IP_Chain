package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)
type Handler struct { db *sql.DB; cache *redis.Client }
func NewHandler(db *sql.DB, cache *redis.Client) *Handler { return &Handler{db:db,cache:cache} }
func (h *Handler) Live(c *gin.Context) { httpx.OK(c,"SERVICE_LIVE",gin.H{"status":"ok"}) }
func (h *Handler) Ready(c *gin.Context) { ctx,cancel:=context.WithTimeout(c.Request.Context(),2*time.Second); defer cancel(); if err:=h.db.PingContext(ctx); err!=nil { httpx.Error(c,http.StatusServiceUnavailable,"MYSQL_UNAVAILABLE","mysql is unavailable"); return }; if err:=h.cache.Ping(ctx).Err(); err!=nil { httpx.Error(c,http.StatusServiceUnavailable,"REDIS_UNAVAILABLE","redis is unavailable"); return }; httpx.OK(c,"SERVICE_READY",gin.H{"status":"ok"}) }
