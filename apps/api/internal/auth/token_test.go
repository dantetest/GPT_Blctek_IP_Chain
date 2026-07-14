package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("test-issuer", "01234567890123456789012345678901", 15*time.Minute)
	user := User{ID: "01TESTUSER00000000000000000", Tier: "PROFESSIONAL", GlobalRole: "USER"}
	raw, expiresAt, err := manager.IssueAccess(user, "01TESTSESSION000000000000000")
	require.NoError(t, err)
	require.True(t, expiresAt.After(time.Now()))
	claims, err := manager.ParseAccess(raw)
	require.NoError(t, err)
	require.Equal(t, user.ID, claims.Subject)
	require.Equal(t, "PROFESSIONAL", claims.Tier)
}

func TestOpaqueTokenHashIsStable(t *testing.T) {
	raw, hash, err := NewOpaqueToken(32)
	require.NoError(t, err)
	require.NotEmpty(t, raw)
	require.Equal(t, hash, HashOpaqueToken(raw))
}
