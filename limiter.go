package ratelimit

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type entry struct {
	count   uint64
	resetAt time.Time
}

type shard struct {
	mu      sync.Mutex
	entries map[[sha256.Size]byte]entry
}

// Limiter is a concurrency-safe, in-memory fixed-window rate limiter.
type Limiter struct {
	config normalizedConfig
	clock  clock
	shards []shard

	keyCount atomic.Uint64
	closed   atomic.Bool
	stopOnce sync.Once
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// New creates a Limiter. Call Close when the limiter is no longer needed.
func New(config Config) (*Limiter, error) {
	return newWithClock(config, realClock{})
}

func newWithClock(config Config, clk clock) (*Limiter, error) {
	normalized, err := config.normalize()
	if err != nil {
		return nil, err
	}

	l := &Limiter{
		config: normalized,
		clock:  clk,
		shards: make([]shard, normalized.shards),
		stopCh: make(chan struct{}),
	}
	for i := range l.shards {
		l.shards[i].entries = make(map[[sha256.Size]byte]entry)
	}

	if !normalized.disableBackgroundCleanup {
		l.wg.Add(1)
		go l.runCleanup()
	}

	return l, nil
}

// Allow records one unit of cost for key and returns the resulting decision.
func (l *Limiter) Allow(key string) (Decision, error) {
	return l.AllowN(key, 1)
}

// AllowN records cost units for key and returns the resulting decision.
// Blocked attempts are included in Count so callers can observe sustained pressure.
func (l *Limiter) AllowN(key string, cost uint64) (Decision, error) {
	if key == "" {
		return Decision{}, ErrEmptyKey
	}
	if cost == 0 {
		return Decision{}, ErrInvalidCost
	}
	if l.closed.Load() {
		return Decision{}, ErrClosed
	}

	hash := sha256.Sum256([]byte(key))
	now := l.clock.now()

	s := l.shardFor(hash)
	s.mu.Lock()
	defer s.mu.Unlock()

	if l.closed.Load() {
		return Decision{}, ErrClosed
	}

	current, ok := s.entries[hash]
	if ok {
		if !now.Before(current.resetAt) {
			current = entry{resetAt: now.Add(l.config.window)}
		}
		current.count = saturatingAdd(current.count, cost)
		s.entries[hash] = current
		return l.decision(current, now), nil
	}

	if !l.reserveKey() {
		return Decision{}, ErrCapacity
	}

	current = entry{
		count:   cost,
		resetAt: now.Add(l.config.window),
	}
	s.entries[hash] = current
	return l.decision(current, now), nil
}

// Reset removes key from the limiter. It reports whether the key existed.
func (l *Limiter) Reset(key string) bool {
	if key == "" {
		return false
	}

	hash := sha256.Sum256([]byte(key))
	s := l.shardFor(hash)
	s.mu.Lock()
	_, ok := s.entries[hash]
	if ok {
		delete(s.entries, hash)
		l.keyCount.Add(^uint64(0))
	}
	s.mu.Unlock()
	return ok
}

// Len returns the number of keys currently retained in memory.
func (l *Limiter) Len() uint64 {
	return l.keyCount.Load()
}

// Cleanup removes expired keys and returns the number removed.
func (l *Limiter) Cleanup() int {
	now := l.clock.now()
	removed := 0
	for i := range l.shards {
		s := &l.shards[i]
		s.mu.Lock()
		for key, current := range s.entries {
			if !now.Before(current.resetAt) {
				delete(s.entries, key)
				removed++
			}
		}
		s.mu.Unlock()
	}
	if removed > 0 {
		l.keyCount.Add(^uint64(removed - 1))
	}
	return removed
}

// Close stops background cleanup. It is safe to call multiple times.
// Calls already in progress may complete; later calls return ErrClosed.
func (l *Limiter) Close() error {
	l.stopOnce.Do(func() {
		l.closed.Store(true)
		close(l.stopCh)
		l.wg.Wait()
	})
	return nil
}

func (l *Limiter) runCleanup() {
	defer l.wg.Done()
	ticker := l.clock.newTicker(l.config.cleanupInterval)
	defer ticker.stop()

	for {
		select {
		case <-ticker.channel():
			l.Cleanup()
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) shardFor(hash [sha256.Size]byte) *shard {
	index := binary.LittleEndian.Uint64(hash[:8]) % uint64(len(l.shards))
	return &l.shards[index]
}

func (l *Limiter) reserveKey() bool {
	for {
		current := l.keyCount.Load()
		if current >= l.config.maxKeys {
			return false
		}
		if l.keyCount.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func (l *Limiter) decision(current entry, now time.Time) Decision {
	remaining := uint64(0)
	if current.count < l.config.limit {
		remaining = l.config.limit - current.count
	}
	allowed := current.count <= l.config.limit
	retryAfter := time.Duration(0)
	if !allowed {
		retryAfter = current.resetAt.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
	}
	return Decision{
		Allowed:    allowed,
		Limit:      l.config.limit,
		Count:      current.count,
		Remaining:  remaining,
		ResetAt:    current.resetAt,
		RetryAfter: retryAfter,
	}
}

func saturatingAdd(left, right uint64) uint64 {
	if math.MaxUint64-left < right {
		return math.MaxUint64
	}
	return left + right
}
