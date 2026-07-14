package principal

import "github.com/gin-gonic/gin"

const contextKey = "principal"

type Value struct {
	UserID    string
	SessionID string
	APIKeyID  string
	AuthType  string
	Role      string
	Tier      string
	Scopes    []string
}

func Set(c *gin.Context, value Value) {
	c.Set(contextKey, value)
}

func Get(c *gin.Context) (Value, bool) {
	value, exists := c.Get(contextKey)
	if !exists {
		return Value{}, false
	}
	result, ok := value.(Value)
	return result, ok
}

func HasScope(value Value, required string) bool {
	for _, scope := range value.Scopes {
		if scope == required {
			return true
		}
	}
	return false
}
