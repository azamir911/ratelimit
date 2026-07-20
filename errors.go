package ratelimit

import "errors"

var (
	// ErrClosed is returned after the limiter has been closed.
	ErrClosed = errors.New("ratelimit: limiter is closed")
	// ErrCapacity is returned when a new key cannot be admitted because MaxKeys is reached.
	ErrCapacity = errors.New("ratelimit: key capacity reached")
	// ErrEmptyKey is returned when an empty key is supplied.
	ErrEmptyKey = errors.New("ratelimit: key must not be empty")
	// ErrInvalidCost is returned when AllowN is called with a zero cost.
	ErrInvalidCost = errors.New("ratelimit: cost must be greater than zero")
)
