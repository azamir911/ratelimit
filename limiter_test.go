package ratelimit

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeTicker struct {
	ch chan time.Time
}

func (t *fakeTicker) channel() <-chan time.Time { return t.ch }
func (t *fakeTicker) stop()                     {}

type fakeClock struct {
	mu      sync.Mutex
	current time.Time
	tickers []*fakeTicker
}

func newFakeClock(start time.Time) *fakeClock {
	return &fakeClock{current: start}
}

func (c *fakeClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

func (c *fakeClock) newTicker(time.Duration) ticker {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &fakeTicker{ch: make(chan time.Time, 1)}
	c.tickers = append(c.tickers, t)
	return t
}

func (c *fakeClock) advance(duration time.Duration) {
	c.mu.Lock()
	c.current = c.current.Add(duration)
	current := c.current
	tickers := append([]*fakeTicker(nil), c.tickers...)
	c.mu.Unlock()

	for _, t := range tickers {
		select {
		case t.ch <- current:
		default:
		}
	}
}

func TestConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
	}{
		{name: "zero limit", config: Config{Window: time.Second}},
		{name: "zero window", config: Config{Limit: 1}},
		{name: "negative shards", config: Config{Limit: 1, Window: time.Second, Shards: -1}},
		{name: "negative cleanup", config: Config{Limit: 1, Window: time.Second, CleanupInterval: -1}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := New(test.config); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestAllowUsesStandardLimitBoundary(t *testing.T) {
	t.Parallel()
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	limiter, err := newWithClock(Config{
		Limit:                    2,
		Window:                   time.Minute,
		MaxKeys:                  10,
		DisableBackgroundCleanup: true,
	}, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	first, err := limiter.Allow("key")
	if err != nil || !first.Allowed || first.Count != 1 || first.Remaining != 1 {
		t.Fatalf("unexpected first decision: %+v, err=%v", first, err)
	}
	second, err := limiter.Allow("key")
	if err != nil || !second.Allowed || second.Count != 2 || second.Remaining != 0 {
		t.Fatalf("unexpected second decision: %+v, err=%v", second, err)
	}
	third, err := limiter.Allow("key")
	if err != nil || third.Allowed || third.Count != 3 || third.RetryAfter != time.Minute {
		t.Fatalf("unexpected third decision: %+v, err=%v", third, err)
	}
}

func TestWindowResetsPerKey(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(start)
	limiter, err := newWithClock(Config{
		Limit:                    1,
		Window:                   time.Minute,
		MaxKeys:                  10,
		DisableBackgroundCleanup: true,
	}, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	first, _ := limiter.Allow("a")
	clock.advance(30 * time.Second)
	other, _ := limiter.Allow("b")
	clock.advance(30 * time.Second)
	reset, _ := limiter.Allow("a")
	blockedOther, _ := limiter.Allow("b")

	if !first.Allowed || !other.Allowed || !reset.Allowed || reset.Count != 1 {
		t.Fatalf("unexpected decisions: first=%+v other=%+v reset=%+v", first, other, reset)
	}
	if blockedOther.Allowed || blockedOther.RetryAfter != 30*time.Second {
		t.Fatalf("unexpected second key decision: %+v", blockedOther)
	}
}

func TestCapacityAndCleanup(t *testing.T) {
	t.Parallel()
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	limiter, err := newWithClock(Config{
		Limit:                    1,
		Window:                   time.Minute,
		MaxKeys:                  2,
		Shards:                   2,
		DisableBackgroundCleanup: true,
	}, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	if _, err := limiter.Allow("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := limiter.Allow("b"); err != nil {
		t.Fatal(err)
	}
	if _, err := limiter.Allow("c"); !errors.Is(err, ErrCapacity) {
		t.Fatalf("expected ErrCapacity, got %v", err)
	}

	clock.advance(time.Minute)
	if removed := limiter.Cleanup(); removed != 2 {
		t.Fatalf("expected two expired keys to be removed, got %d", removed)
	}
	if _, err := limiter.Allow("c"); err != nil {
		t.Fatalf("expected explicit cleanup to admit key: %v", err)
	}
	if got := limiter.Len(); got != 1 {
		t.Fatalf("expected one retained key, got %d", got)
	}
}

func TestBackgroundCleanup(t *testing.T) {
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	limiter, err := newWithClock(Config{
		Limit:           1,
		Window:          time.Minute,
		MaxKeys:         10,
		CleanupInterval: time.Second,
	}, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	if _, err := limiter.Allow("a"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(time.Second)
	for {
		clock.mu.Lock()
		ready := len(clock.tickers) > 0
		clock.mu.Unlock()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cleanup ticker was not created")
		}
		time.Sleep(time.Millisecond)
	}

	clock.advance(time.Minute)
	deadline = time.Now().Add(time.Second)
	for limiter.Len() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("expected background cleanup to remove key, len=%d", limiter.Len())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestReset(t *testing.T) {
	t.Parallel()
	limiter, err := New(Config{Limit: 1, Window: time.Minute, MaxKeys: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	if _, err := limiter.Allow("a"); err != nil {
		t.Fatal(err)
	}
	if !limiter.Reset("a") || limiter.Reset("a") {
		t.Fatal("unexpected reset result")
	}
	if got := limiter.Len(); got != 0 {
		t.Fatalf("expected zero keys, got %d", got)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()
	limiter, err := New(Config{Limit: 1, Window: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if err := limiter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := limiter.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := limiter.Allow("a"); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestInputErrors(t *testing.T) {
	t.Parallel()
	limiter, err := New(Config{Limit: 1, Window: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	if _, err := limiter.Allow(""); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("expected ErrEmptyKey, got %v", err)
	}
	if _, err := limiter.AllowN("a", 0); !errors.Is(err, ErrInvalidCost) {
		t.Fatalf("expected ErrInvalidCost, got %v", err)
	}
}
