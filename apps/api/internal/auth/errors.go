package auth

import "errors"

var (
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrEmailNotVerified  = errors.New("email is not verified")
	ErrAccountLocked     = errors.New("account is locked")
	ErrAccountDisabled   = errors.New("account is disabled")
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("expired token")
	ErrWeakPassword      = errors.New("password does not meet policy")
)
