package apikey

import "errors"

var (
	ErrNotAvailable     = errors.New("api keys are not available for this tier")
	ErrLimitReached     = errors.New("api key limit reached")
	ErrInvalidName      = errors.New("api key name is invalid")
	ErrInvalidScope     = errors.New("api key scope is invalid")
	ErrInvalidKey       = errors.New("api key is invalid")
	ErrQuotaExceeded    = errors.New("api key monthly quota exceeded")
	ErrKeyNotFound      = errors.New("api key not found")
	ErrAccountNotActive = errors.New("api key owner is not active")
)
