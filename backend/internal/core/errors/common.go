// Package core_errors defines standardized sentinel errors used across all application layers
// for consistent HTTP mapping, domain error handling, and testing.
package core_errors

import "errors"

var (
	// ErrNotFound indicates that a requested entity (user, note, dish, workout) does not exist or was deleted.
	// Maps to HTTP 404 Not Found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidArgument indicates malformed request payload, invalid parameters, or validation failure.
	// Maps to HTTP 400 Bad Request.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrConflict indicates an optimistic lock version conflict (e.g., outdated entity version).
	// Maps to HTTP 409 Conflict.
	ErrConflict = errors.New("conflict")

	// ErrUnauthorized indicates missing, invalid, or expired credentials (JWT token).
	// Maps to HTTP 401 Unauthorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTooManyRequests indicates that a client exceeded the allowed rate limit threshold.
	// Maps to HTTP 429 Too Many Requests.
	ErrTooManyRequests = errors.New("too many requests")
)
