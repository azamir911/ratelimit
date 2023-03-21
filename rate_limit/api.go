package rate_limit

import "net/url"

type RateLimiter interface {
	Allow(*url.URL) (int, bool, error)
	Stop()
}

func New(threshold int, ttl int) RateLimiter {
	return newRateLimiter(threshold, ttl)
}
