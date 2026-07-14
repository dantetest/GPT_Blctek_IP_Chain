package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := HashPassword("CorrectHorse9Battery")
	require.NoError(t, err)
	require.True(t, VerifyPassword("CorrectHorse9Battery", hash))
	require.False(t, VerifyPassword("WrongPassword9", hash))
}

func TestPasswordPolicy(t *testing.T) {
	require.ErrorIs(t, ValidatePassword("short1A"), ErrWeakPassword)
	require.ErrorIs(t, ValidatePassword("alllowercase123"), ErrWeakPassword)
	require.NoError(t, ValidatePassword("LongEnoughPassword9"))
}
