package idempotency

import "errors"

var (
	ErrKeyRequired     = errors.New("idempotency key is required")
	ErrKeyInvalid      = errors.New("idempotency key is invalid")
	ErrRequestConflict = errors.New("idempotency key was used for another request")
	ErrRequestInFlight = errors.New("idempotent request is still in progress")
	ErrRequestTooLarge = errors.New("request body exceeds idempotency limit")
)
