package ratelimit

import (
	"fmt"
	"time"
)

const (
	defaultShards  = 64
	defaultMaxKeys = 100_000
)

// Config controls a fixed-window Limiter.
type Config struct {
	// Limit is the maximum accumulated cost allowed for one key in a window.
	Limit uint64
	// Window is the duration of each per-key fixed window.
	Window time.Duration
	// MaxKeys bounds the number of keys retained in memory. Zero uses 100,000.
	MaxKeys uint64
	// Shards controls lock striping. Zero uses 64.
	Shards int
	// CleanupInterval controls background removal of expired keys. Zero uses Window.
	CleanupInterval time.Duration
	// DisableBackgroundCleanup disables the cleanup goroutine. Expired keys are still
	// reset lazily when observed, and Cleanup can be called explicitly.
	DisableBackgroundCleanup bool
}

type normalizedConfig struct {
	limit                    uint64
	window                   time.Duration
	maxKeys                  uint64
	shards                   int
	cleanupInterval          time.Duration
	disableBackgroundCleanup bool
}

func (c Config) normalize() (normalizedConfig, error) {
	if c.Limit == 0 {
		return normalizedConfig{}, fmt.Errorf("ratelimit: limit must be greater than zero")
	}
	if c.Window <= 0 {
		return normalizedConfig{}, fmt.Errorf("ratelimit: window must be greater than zero")
	}

	maxKeys := c.MaxKeys
	if maxKeys == 0 {
		maxKeys = defaultMaxKeys
	}

	shards := c.Shards
	if shards == 0 {
		shards = defaultShards
	}
	if shards < 1 {
		return normalizedConfig{}, fmt.Errorf("ratelimit: shards must be greater than zero")
	}
	if uint64(shards) > maxKeys {
		shards = int(maxKeys)
	}

	cleanupInterval := c.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = c.Window
	}
	if cleanupInterval < 0 {
		return normalizedConfig{}, fmt.Errorf("ratelimit: cleanup interval must not be negative")
	}

	return normalizedConfig{
		limit:                    c.Limit,
		window:                   c.Window,
		maxKeys:                  maxKeys,
		shards:                   shards,
		cleanupInterval:          cleanupInterval,
		disableBackgroundCleanup: c.DisableBackgroundCleanup,
	}, nil
}
