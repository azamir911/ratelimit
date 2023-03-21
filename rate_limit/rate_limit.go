package rate_limit

import (
	"errors"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"net/url"
	"sync"
	"time"
)

// tickerDuration declared the duration to run the ticker with
var tickerDuration = 5 * time.Minute

//var tickerDuration = 10 * time.Second

func newRateLimiter(threshold int, ttl int) RateLimiter {
	ctx := context.WithValue(context.Background(), "RateLimit", "True")
	r := &rateLimiterCounter{
		counter:   make(map[uuid.UUID]*rateLimitValue),
		mut:       sync.RWMutex{},
		threshold: threshold,
		ttl:       ttl,
		ctx:       ctx,
	}

	r.ticker = time.NewTicker(tickerDuration)
	// The ticker will run in a different goroutine.
	// To stop it use rateLimit.Stop()
	go runTicker(r)

	return r
}

type rateLimiterCounter struct {
	// counter is a map that holds all the requests its attributes
	counter map[uuid.UUID]*rateLimitValue
	mut     sync.RWMutex
	// threshold from arguments
	threshold int
	// TTL from arguments
	ttl    int
	ticker *time.Ticker
	// The rate limit context
	ctx context.Context
}

type rateLimitValue struct {
	ttl     time.Time
	counter int
}

// runTicker should run in a different goroutine and responsible to clean the rate limit
func runTicker(r *rateLimiterCounter) {
	for {
		select {
		case <-r.ticker.C:
			r.clean()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *rateLimiterCounter) Allow(url *url.URL) (int, bool, error) {
	if url == nil {
		return 0, false, errors.New("nil url")
	}

	key, err := newMd5UUID("", url.Path)
	if err != nil {
		return 0, false, err
	}

	r.mut.Lock()
	defer r.mut.Unlock()

	value, ok := r.counter[key]
	utcTime := time.Now().UTC()
	if !ok || r.isReachTtl(*value, utcTime) {
		// In case it's the first time for URL, or it already reaches the TTL, a new rateLimitValue will created
		value = &rateLimitValue{
			ttl:     utcTime.Add(time.Millisecond * time.Duration(r.ttl)),
			counter: 1,
		}
		r.counter[key] = value
	} else {
		// Increasing the counter
		value.counter++
	}

	// Validate if the URL is reach the threshold
	isReachCounter := r.isCounterReachThreshold(*value)

	return value.counter, !isReachCounter, nil
}

func (r *rateLimiterCounter) Stop() {
	r.ctx.Done()
}

// isReachTtl return true in case the TTL has passed
func (r *rateLimiterCounter) isReachTtl(value rateLimitValue, utcTime time.Time) bool {
	return utcTime.After(value.ttl)
}

// isCounterReachThreshold return true in case the counter reach the threshold
func (r *rateLimiterCounter) isCounterReachThreshold(value rateLimitValue) bool {
	return value.counter > r.threshold
}

// clean removing old elements from the counter map. It makes sure to clean unused URLs
func (r *rateLimiterCounter) clean() {
	utcTime := time.Now().UTC()
	for key, value := range r.counter {
		if r.isReachTtl(*value, utcTime) {
			r.mut.Lock()
			delete(r.counter, key)
			r.mut.Unlock()
		}
	}
}
