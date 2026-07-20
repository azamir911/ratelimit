package ratelimit_test

import (
	"fmt"
	"time"

	"github.com/azamir911/ratelimit"
)

func Example() {
	limiter, err := ratelimit.New(ratelimit.Config{
		Limit:   2,
		Window:  time.Minute,
		MaxKeys: 10_000,
	})
	if err != nil {
		panic(err)
	}
	defer limiter.Close()

	first, _ := limiter.Allow("tenant:42")
	second, _ := limiter.Allow("tenant:42")
	third, _ := limiter.Allow("tenant:42")

	fmt.Println(first.Allowed, first.Remaining)
	fmt.Println(second.Allowed, second.Remaining)
	fmt.Println(third.Allowed, third.Remaining)
	// Output:
	// true 1
	// true 0
	// false 0
}
