package middleware

import (
	"net/http"
	"strings"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/auth"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/httpx"
	"github.com/gin-gonic/gin"
)

const principalKey = "principal"

type Principal struct {
	UserID    string
	SessionID string
	Role      string
	Tier      string
}

func Authenticate(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "a bearer access token is required")
			return
		}
		claims, err := tokens.ParseAccess(strings.TrimSpace(parts[1]))
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "INVALID_ACCESS_TOKEN", "access token is invalid or expired")
			return
		}
		c.Set(principalKey, Principal{
			UserID:    claims.Subject,
			SessionID: claims.SessionID,
			Role:      claims.Role,
			Tier:      claims.Tier,
		})
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		principal, ok := CurrentPrincipal(c)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
			return
		}
		if _, ok := allowed[principal.Role]; !ok {
			httpx.Error(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
			return
		}
		c.Next()
	}
}

func CurrentPrincipal(c *gin.Context) (Principal, bool) {
	value, exists := c.Get(principalKey)
	if !exists {
		return Principal{}, false
	}
	principal, ok := value.(Principal)
	return principal, ok
}
