package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/id"
)

type ServiceConfig struct {
	RefreshTokenTTL  time.Duration
	AutoVerifyEmail  bool
	LoginMaxAttempts int
	LoginLockTTL     time.Duration
}

type Service struct {
	repository *Repository
	tokens     *TokenManager
	config     ServiceConfig
	now        func() time.Time
}

func NewService(repository *Repository, tokens *TokenManager, config ServiceConfig) *Service {
	return &Service{repository: repository, tokens: tokens, config: config, now: time.Now}
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type RefreshInput struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (UserView, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return UserView{}, err
	}
	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return UserView{}, err
	}
	userID, err := id.New()
	if err != nil {
		return UserView{}, err
	}
	now := s.now().UTC()
	status := "PENDING_EMAIL"
	var verifiedAt *time.Time
	if s.config.AutoVerifyEmail {
		status = "ACTIVE"
		verifiedAt = &now
	}
	user := User{
		ID:              userID,
		Email:           email,
		PasswordHash:    passwordHash,
		DisplayName:     strings.TrimSpace(input.DisplayName),
		Status:          status,
		Tier:            "BASIC",
		GlobalRole:      "USER",
		EmailVerifiedAt: verifiedAt,
		CreatedAt:       now,
	}

	verificationToken, verificationHash, err := NewOpaqueToken(32)
	if err != nil {
		return UserView{}, err
	}
	verificationID, err := id.New()
	if err != nil {
		return UserView{}, err
	}
	outboxID, err := id.New()
	if err != nil {
		return UserView{}, err
	}
	err = s.repository.CreateUser(ctx, CreateUserParams{
		User:               user,
		VerificationID:     verificationID,
		VerificationHash:   verificationHash,
		VerificationToken:  verificationToken,
		VerificationExpiry: now.Add(30 * time.Minute),
		OutboxEventID:      outboxID,
	}, s.config.AutoVerifyEmail)
	if err != nil {
		return UserView{}, err
	}
	return user.View(), nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (TokenPair, UserView, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return TokenPair{}, UserView{}, ErrInvalidCredential
	}
	user, err := s.repository.FindUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return TokenPair{}, UserView{}, ErrInvalidCredential
	}
	if err != nil {
		return TokenPair{}, UserView{}, fmt.Errorf("find login user: %w", err)
	}

	now := s.now().UTC()
	if user.Status == "DISABLED" {
		return TokenPair{}, UserView{}, ErrAccountDisabled
	}
	if user.Status == "PENDING_EMAIL" {
		return TokenPair{}, UserView{}, ErrEmailNotVerified
	}
	if user.Status == "LOCKED" {
		if user.LockedUntil != nil && user.LockedUntil.After(now) {
			return TokenPair{}, UserView{}, ErrAccountLocked
		}
		if err := s.repository.UnlockUser(ctx, user.ID); err != nil {
			return TokenPair{}, UserView{}, fmt.Errorf("unlock user: %w", err)
		}
		user.Status = "ACTIVE"
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
	}
	if !VerifyPassword(input.Password, user.PasswordHash) {
		lockUntil := now.Add(s.config.LoginLockTTL)
		if err := s.repository.RecordLoginFailure(ctx, user, s.config.LoginMaxAttempts, lockUntil); err != nil {
			return TokenPair{}, UserView{}, fmt.Errorf("record login failure: %w", err)
		}
		return TokenPair{}, UserView{}, ErrInvalidCredential
	}

	pair, session, err := s.newSession(user, input.UserAgent, input.IPAddress)
	if err != nil {
		return TokenPair{}, UserView{}, err
	}
	if err := s.repository.CreateSession(ctx, session); err != nil {
		return TokenPair{}, UserView{}, fmt.Errorf("create auth session: %w", err)
	}
	if err := s.repository.RecordLoginSuccess(ctx, user.ID, now); err != nil {
		return TokenPair{}, UserView{}, fmt.Errorf("record login success: %w", err)
	}
	return pair, user.View(), nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (TokenPair, error) {
	if input.RefreshToken == "" {
		return TokenPair{}, ErrInvalidToken
	}
	session, user, err := s.repository.FindSessionWithUser(ctx, HashOpaqueToken(input.RefreshToken))
	if errors.Is(err, sql.ErrNoRows) {
		return TokenPair{}, ErrInvalidToken
	}
	if err != nil {
		return TokenPair{}, fmt.Errorf("find refresh session: %w", err)
	}
	now := s.now().UTC()
	if session.RevokedAt != nil {
		return TokenPair{}, ErrInvalidToken
	}
	if !session.ExpiresAt.After(now) {
		return TokenPair{}, ErrExpiredToken
	}
	if user.Status != "ACTIVE" {
		return TokenPair{}, ErrAccountDisabled
	}

	pair, newSession, err := s.newSession(user, input.UserAgent, input.IPAddress)
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.repository.RotateSession(ctx, session.ID, newSession); err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.repository.RevokeSessionByRefreshHash(ctx, HashOpaqueToken(refreshToken))
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return ErrInvalidToken
	}
	return s.repository.VerifyEmail(ctx, HashOpaqueToken(rawToken), s.now().UTC())
}

func (s *Service) GetUser(ctx context.Context, userID string) (UserView, error) {
	user, err := s.repository.FindUserByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserView{}, ErrInvalidToken
	}
	if err != nil {
		return UserView{}, err
	}
	return user.View(), nil
}

func (s *Service) newSession(user User, userAgent, ipAddress string) (TokenPair, Session, error) {
	sessionID, err := id.New()
	if err != nil {
		return TokenPair{}, Session{}, err
	}
	refreshToken, refreshHash, err := NewOpaqueToken(48)
	if err != nil {
		return TokenPair{}, Session{}, err
	}
	accessToken, accessExpiresAt, err := s.tokens.IssueAccess(user, sessionID)
	if err != nil {
		return TokenPair{}, Session{}, err
	}
	session := Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        truncate(userAgent, 512),
		IPAddress:        truncate(ipAddress, 45),
		ExpiresAt:        s.now().UTC().Add(s.config.RefreshTokenTTL),
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    accessExpiresAt,
	}, session, nil
}

func normalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(normalized)
	if err != nil || address.Address != normalized || len(normalized) > 320 {
		return "", ErrInvalidEmail
	}
	return normalized, nil
}

func truncate(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	return value[:maximum]
}
