package auth

import "time"

type User struct {
	ID                  string
	Email               string
	PasswordHash        string
	DisplayName         string
	Status              string
	Tier                string
	GlobalRole          string
	EmailVerifiedAt     *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	CreatedAt           time.Time
}

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserView struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	Status          string     `json:"status"`
	Tier            string     `json:"tier"`
	GlobalRole      string     `json:"global_role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (u User) View() UserView {
	return UserView{
		ID:              u.ID,
		Email:           u.Email,
		DisplayName:     u.DisplayName,
		Status:          u.Status,
		Tier:            u.Tier,
		GlobalRole:      u.GlobalRole,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}
