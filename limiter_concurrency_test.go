package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentSameKeyHasExactCount(t *testing.T) {
	limiter, err := New(Config{
		Limit:   10_000,
		Window:  time.Minute,
		MaxKeys: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	const goroutines = 64
	const callsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				if _, err := limiter.Allow("shared"); err != nil {
					t.Errorf("allow: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	decision, err := limiter.Allow("shared")
	if err != nil {
		t.Fatal(err)
	}
	want := uint64(goroutines*callsPerGoroutine + 1)
	if decision.Count != want {
		t.Fatalf("expected count %d, got %d", want, decision.Count)
	}
}

func TestConcurrentDistinctKeysRespectCapacity(t *testing.T) {
	const maxKeys = 500
	limiter, err := New(Config{
		Limit:   1,
		Window:  time.Minute,
		MaxKeys: maxKeys,
		Shards:  32,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	var wg sync.WaitGroup
	for i := 0; i < maxKeys*2; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = limiter.Allow(string(rune(i + 1)))
		}()
	}
	wg.Wait()

	if got := limiter.Len(); got != maxKeys {
		t.Fatalf("expected %d keys, got %d", maxKeys, got)
	}
}
