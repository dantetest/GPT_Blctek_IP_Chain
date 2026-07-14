package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/apikey"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/httpx"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/principal"
	"github.com/gin-gonic/gin"
)

func AuthenticateAPIKey(service *apikey.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if raw == "" {
			httpx.Error(c, http.StatusUnauthorized, "API_KEY_REQUIRED", "X-API-Key is required")
			return
		}
		keyPrincipal, err := service.Authenticate(c.Request.Context(), raw)
		if err != nil {
			switch {
			case errors.Is(err, apikey.ErrQuotaExceeded):
				httpx.Error(c, http.StatusTooManyRequests, "API_KEY_QUOTA_EXCEEDED", "api key monthly quota exceeded")
			case errors.Is(err, apikey.ErrAccountNotActive):
				httpx.Error(c, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE", "api key owner account is not active")
			default:
				httpx.Error(c, http.StatusUnauthorized, "INVALID_API_KEY", "api key is invalid or expired")
			}
			return
		}
		principal.Set(c, principal.Value{
			UserID:   keyPrincipal.UserID,
			APIKeyID: keyPrincipal.KeyID,
			AuthType: "API_KEY",
			Role:     keyPrincipal.Role,
			Tier:     keyPrincipal.Tier,
			Scopes:   keyPrincipal.Scopes,
		})
		c.Next()
	}
}

func RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, ok := principal.Get(c)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
			return
		}
		for _, scope := range scopes {
			if !principal.HasScope(current, scope) {
				httpx.Error(c, http.StatusForbidden, "INSUFFICIENT_SCOPE", "api key does not grant the required scope")
				return
			}
		}
		c.Next()
	}
}
