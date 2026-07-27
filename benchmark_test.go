package ratelimit

import (
	"strconv"
	"testing"
	"time"
)

func BenchmarkAllowSameKey(b *testing.B) {
	limiter, err := New(Config{
		Limit:                    ^uint64(0),
		Window:                   time.Hour,
		MaxKeys:                  1,
		DisableBackgroundCleanup: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer limiter.Close()

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := limiter.Allow("shared"); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAllowDistinctKeys(b *testing.B) {
	limiter, err := New(Config{
		Limit:                    1,
		Window:                   time.Hour,
		MaxKeys:                  uint64(b.N) + 1,
		DisableBackgroundCleanup: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer limiter.Close()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow(strconv.Itoa(i)); err != nil {
			b.Fatal(err)
		}
	}
}
