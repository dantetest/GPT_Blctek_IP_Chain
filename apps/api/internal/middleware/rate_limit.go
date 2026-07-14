package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimitKey func(*gin.Context) string

func RateLimit(client *redis.Client, namespace string, limit int64, window time.Duration, key RateLimitKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := key(c)
		bucket := time.Now().UTC().Unix() / int64(window.Seconds())
		redisKey := fmt.Sprintf("rate:%s:%s:%d", namespace, identifier, bucket)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		count, err := client.Incr(ctx, redisKey).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			client.Expire(ctx, redisKey, window+time.Second)
		}
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		if count > limit {
			c.Header("Retry-After", strconv.FormatInt(int64(window.Seconds()), 10))
			httpx.Error(c, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
			return
		}
		c.Next()
	}
}

func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}
