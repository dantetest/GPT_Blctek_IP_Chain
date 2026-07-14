package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/id"
	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	SessionID string `json:"sid"`
	Role      string `json:"role"`
	Tier      string `json:"tier"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	issuer string
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(issuer, secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{issuer: issuer, secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *TokenManager) IssueAccess(user User, sessionID string) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.ttl)
	tokenID, err := id.New()
	if err != nil {
		return "", time.Time{}, err
	}
	claims := AccessClaims{
		SessionID: sessionID,
		Role:      user.GlobalRole,
		Tier:      user.Tier,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{"blctekip-api"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *TokenManager) ParseAccess(raw string) (AccessClaims, error) {
	claims := AccessClaims{}
	token, err := jwt.ParseWithClaims(
		raw,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience("blctekip-api"),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return AccessClaims{}, ErrExpiredToken
		}
		return AccessClaims{}, ErrInvalidToken
	}
	if !token.Valid || claims.Subject == "" || claims.SessionID == "" {
		return AccessClaims{}, ErrInvalidToken
	}
	return claims, nil
}

func NewOpaqueToken(bytesLength int) (string, string, error) {
	value := make([]byte, bytesLength)
	if _, err := rand.Read(value); err != nil {
		return "", "", fmt.Errorf("generate opaque token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(value)
	return raw, HashOpaqueToken(raw), nil
}

func HashOpaqueToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
