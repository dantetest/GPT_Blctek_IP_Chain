package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const keyPrefix = "bkip"

func NewCredential() (raw, prefix, hash string, err error) {
	prefixBytes := make([]byte, 6)
	secretBytes := make([]byte, 32)
	if _, err = rand.Read(prefixBytes); err != nil {
		return "", "", "", fmt.Errorf("generate api key prefix: %w", err)
	}
	if _, err = rand.Read(secretBytes); err != nil {
		return "", "", "", fmt.Errorf("generate api key secret: %w", err)
	}
	prefix = strings.ToLower(hex.EncodeToString(prefixBytes))
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)
	raw = keyPrefix + "_" + prefix + "_" + secret
	return raw, prefix, Hash(raw), nil
}

func Parse(raw string) (string, error) {
	parts := strings.Split(raw, "_")
	if len(parts) != 3 || parts[0] != keyPrefix || len(parts[1]) != 12 || len(parts[2]) < 40 {
		return "", ErrInvalidKey
	}
	return strings.ToLower(parts[1]), nil
}

func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func NormalizeScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, ErrInvalidScope
	}
	seen := make(map[string]struct{}, len(scopes))
	result := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if _, allowed := allowedScopes[scope]; !allowed {
			return nil, ErrInvalidScope
		}
		if _, duplicate := seen[scope]; duplicate {
			continue
		}
		seen[scope] = struct{}{}
		result = append(result, scope)
	}
	return result, nil
}
