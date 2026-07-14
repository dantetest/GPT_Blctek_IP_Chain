package apikey

import "time"

const (
	StatusActive  = "ACTIVE"
	StatusRevoked = "REVOKED"
)

var allowedScopes = map[string]struct{}{
	"datasets:read":   {},
	"orders:read":     {},
	"orders:write":    {},
	"assets:read":     {},
	"assistant:query": {},
}

type Key struct {
	ID           string
	UserID       string
	Name         string
	KeyPrefix    string
	SecretHash   string
	Scopes       []string
	Status       string
	MonthlyQuota *int64
	MonthlyUsed  int64
	QuotaMonth   time.Time
	ExpiresAt    *time.Time
	LastUsedAt   *time.Time
	CreatedAt    time.Time
	RevokedAt    *time.Time
}

type View struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Scopes       []string   `json:"scopes"`
	Status       string     `json:"status"`
	MonthlyQuota *int64     `json:"monthly_quota,omitempty"`
	MonthlyUsed  int64      `json:"monthly_used"`
	QuotaMonth   string     `json:"quota_month"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

type Created struct {
	View
	APIKey string `json:"api_key"`
}

func (key Key) View() View {
	return View{
		ID:           key.ID,
		Name:         key.Name,
		KeyPrefix:    key.KeyPrefix,
		Scopes:       key.Scopes,
		Status:       key.Status,
		MonthlyQuota: key.MonthlyQuota,
		MonthlyUsed:  key.MonthlyUsed,
		QuotaMonth:   key.QuotaMonth.Format("2006-01"),
		ExpiresAt:    key.ExpiresAt,
		LastUsedAt:   key.LastUsedAt,
		CreatedAt:    key.CreatedAt,
		RevokedAt:    key.RevokedAt,
	}
}

type Principal struct {
	KeyID     string
	UserID    string
	Role      string
	Tier      string
	Scopes    []string
	ExpiresAt *time.Time
}
