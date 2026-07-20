package ratelimit

import "time"

// Decision describes limiter state after an observation.
type Decision struct {
	Allowed    bool
	Limit      uint64
	Count      uint64
	Remaining  uint64
	ResetAt    time.Time
	RetryAfter time.Duration
}
