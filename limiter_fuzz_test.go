package ratelimit

import (
	"testing"
	"time"
)

func FuzzAllow(f *testing.F) {
	f.Add("tenant:42", uint64(1))
	f.Add("https://example.com/a?x=1", uint64(3))
	f.Add("", uint64(0))

	f.Fuzz(func(t *testing.T, key string, cost uint64) {
		limiter, err := New(Config{
			Limit:                    100,
			Window:                   time.Minute,
			MaxKeys:                  10,
			DisableBackgroundCleanup: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer limiter.Close()

		_, _ = limiter.AllowN(key, cost)
	})
}
