package idempotency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashRequestIsStableAndUserScoped(t *testing.T) {
	first := hashRequest("POST", "/api/v1/api-keys", "user-a", []byte(`{"name":"test"}`))
	second := hashRequest("POST", "/api/v1/api-keys", "user-a", []byte(`{"name":"test"}`))
	otherUser := hashRequest("POST", "/api/v1/api-keys", "user-b", []byte(`{"name":"test"}`))
	require.Equal(t, first, second)
	require.NotEqual(t, first, otherUser)
}
