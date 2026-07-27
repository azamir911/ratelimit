package ratelimit

import "time"

// Decision describes the limiter state after an observation.
type Decision struct {
	// Allowed reports whether the accumulated cost is within the configured limit.
	Allowed bool
	// Limit is the configured per-window limit.
	Limit uint64
	// Count is the accumulated cost observed in the current window, including blocked attempts.
	Count uint64
	// Remaining is the cost that can still be accepted in the current window.
	Remaining uint64
	// ResetAt is when the current window expires.
	ResetAt time.Time
	// RetryAfter is non-zero only for a blocked decision.
	RetryAfter time.Duration
}
