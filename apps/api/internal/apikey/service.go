package apikey

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/id"
)

type Service struct {
	repository *Repository
	now        func() time.Time
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

type CreateInput struct {
	UserID    string
	Name      string
	Scopes    []string
	ExpiresAt *time.Time
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Created, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 120 {
		return Created{}, ErrInvalidName
	}
	scopes, err := NormalizeScopes(input.Scopes)
	if err != nil {
		return Created{}, err
	}
	entitlement, err := s.repository.UserEntitlement(ctx, input.UserID)
	if errors.Is(err, sql.ErrNoRows) || entitlement.Status != "ACTIVE" {
		return Created{}, ErrAccountNotActive
	}
	if err != nil {
		return Created{}, err
	}

	limit, quota, allowed := tierPolicy(entitlement.Tier)
	if !allowed {
		return Created{}, ErrNotAvailable
	}
	now := s.now().UTC()
	if input.ExpiresAt != nil && !input.ExpiresAt.After(now) {
		return Created{}, ErrInvalidKey
	}
	raw, prefix, hash, err := NewCredential()
	if err != nil {
		return Created{}, err
	}
	keyID, err := id.New()
	if err != nil {
		return Created{}, err
	}
	key := Key{
		ID:           keyID,
		UserID:       input.UserID,
		Name:         name,
		KeyPrefix:    prefix,
		SecretHash:   hash,
		Scopes:       scopes,
		Status:       StatusActive,
		MonthlyQuota: quota,
		QuotaMonth:   time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:    input.ExpiresAt,
		CreatedAt:    now,
	}
	if err := s.repository.CreateWithLimit(ctx, key, limit); err != nil {
		return Created{}, err
	}
	return Created{View: key.View(), APIKey: raw}, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]View, error) {
	keys, err := s.repository.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(keys))
	for _, key := range keys {
		views = append(views, key.View())
	}
	return views, nil
}

func (s *Service) Authenticate(ctx context.Context, raw string) (Principal, error) {
	return s.repository.AuthenticateAndConsume(ctx, raw, s.now().UTC())
}

func (s *Service) Revoke(ctx context.Context, userID, keyID string) error {
	return s.repository.Revoke(ctx, userID, keyID, s.now().UTC())
}

func tierPolicy(tier string) (limit *int, quota *int64, allowed bool) {
	switch tier {
	case "PROFESSIONAL":
		keyLimit := 5
		monthlyQuota := int64(10000)
		return &keyLimit, &monthlyQuota, true
	case "ENTERPRISE":
		return nil, nil, true
	default:
		return nil, nil, false
	}
}
