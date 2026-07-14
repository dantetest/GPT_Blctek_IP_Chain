package apikey

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialRoundTrip(t *testing.T) {
	raw, prefix, hash, err := NewCredential()
	require.NoError(t, err)
	parsedPrefix, err := Parse(raw)
	require.NoError(t, err)
	require.Equal(t, prefix, parsedPrefix)
	require.Equal(t, hash, Hash(raw))
}

func TestNormalizeScopes(t *testing.T) {
	scopes, err := NormalizeScopes([]string{"datasets:read", "DATASETS:READ", "orders:write"})
	require.NoError(t, err)
	require.Equal(t, []string{"datasets:read", "orders:write"}, scopes)
	require.ErrorIs(t, func() error {
		_, err := NormalizeScopes([]string{"admin:all"})
		return err
	}(), ErrInvalidScope)
}
