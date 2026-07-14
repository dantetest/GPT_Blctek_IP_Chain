package principal

import "github.com/gin-gonic/gin"

const contextKey = "principal"

type Value struct {
	UserID    string
	SessionID string
	Role      string
	Tier      string
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
