package rate_limit

import (
	"errors"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"net/url"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(*url.URL) (int, bool, error)
	State(*url.URL) (bool, error)
	Stop()
}

//var tickerDuration = 5 * time.Minute
var tickerDuration = 10 * time.Second

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
	go runTicker(r)

	return r
}

type rateLimiterCounter struct {
	counter   map[uuid.UUID]*rateLimitValue
	mut       sync.RWMutex
	threshold int
	ttl       int
	ticker    *time.Ticker
	ctx       context.Context
}

type rateLimitValue struct {
	ttl     time.Time
	counter int
}

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

	key, err := NewMd5UUID("", url.Path)
	if err != nil {
		return 0, false, err
	}

	r.mut.Lock()
	defer r.mut.Unlock()

	value, ok := r.counter[key]
	utcTime := time.Now().UTC()
	if !ok || r.isReachTtl(*value, utcTime) {
		//ttl := time.Now().UTC().Add(time.Microsecond * time.Duration(r.ttl))
		//ttl := utcTime.Add(time.Second * 60)
		//ttl := utcTime.Add(time.Millisecond * 60000)
		ttl := utcTime.Add(time.Millisecond * time.Duration(r.ttl))
		value = &rateLimitValue{
			//ttl:     utcTime.Add(time.Duration(r.ttl)),
			ttl:     ttl,
			counter: 1,
		}
		r.counter[key] = value
	} else {
		value.counter++
	}

	isReachCounter := r.isReachCounter(*value)

	return value.counter, !isReachCounter, nil
}

func (r *rateLimiterCounter) State(url *url.URL) (bool, error) {
	if url == nil {
		return false, errors.New("nil url")
	}

	key, err := NewMd5UUID("", url.Path)
	if err != nil {
		return false, err
	}

	r.mut.RLock()
	defer r.mut.RUnlock()

	value, ok := r.counter[key]
	utcTime := time.Now().UTC()
	if !ok || r.isReachTtl(*value, utcTime) {
		return true, nil
	}

	allow := r.isReachCounter(*value)

	return allow, nil
}

func (r *rateLimiterCounter) Stop() {
	r.ctx.Done()
}

func (r *rateLimiterCounter) isReachTtl(value rateLimitValue, utcTime time.Time) bool {
	return utcTime.After(value.ttl)
}

func (r *rateLimiterCounter) isReachCounter(value rateLimitValue) bool {
	return value.counter > r.threshold
}

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
