package rate_limit

import (
	"errors"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"log"
	"net/url"
	"sync"
	"time"
)

// tickerDuration declared the duration to run the ticker with
var tickerDuration = time.Hour

func newRateLimiter(threshold int, ttl int) RateLimiter {
	ctx := context.WithValue(context.Background(), "RateLimit", "True")
	r := &rateLimiter{
		elements:  make(map[uuid.UUID]*value),
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

type rateLimiter struct {
	// elements is a map that holds all the requests its attributes
	elements map[uuid.UUID]*value
	mut      sync.RWMutex
	// mutexes is a map to handle locking for each element
	mutexes sync.Map
	// threshold from arguments
	threshold int
	// ttl from arguments
	ttl    int
	ticker *time.Ticker
	// The rate limit context to stop the ticker
	ctx context.Context
}

type value struct {
	ttl     time.Time
	counter int
}

// runTicker should run in a different goroutine and responsible to clean the rate limit
func runTicker(r *rateLimiter) {
	for {
		select {
		case <-r.ticker.C:
			r.clean()
		case <-r.ctx.Done():
			log.Printf("Ticker context was done")
			return
		}
	}
}

func (r *rateLimiter) Allow(url *url.URL) (int, bool, error) {
	if url == nil {
		return 0, false, errors.New("nil url")
	}

	key, err := newMd5UUID("", url.Path)
	if err != nil {
		return 0, false, err
	}

	// Getting a lock for an element
	elementUnlock := r.elementLock(key)
	// Making sure to release the lock (can be written after the if statement instead, line #94)
	defer elementUnlock()

	v, ok := r.elementsRead(key)
	utcTime := time.Now().UTC()
	if !ok || r.isReachTtl(*v, utcTime) {
		// In case it's the first time for URL, or it already reaches the TTL, a new value will create
		v = &value{
			ttl:     utcTime.Add(time.Millisecond * time.Duration(r.ttl)),
			counter: 1,
		}
		r.elementsStore(key, v)
	} else {
		// Increasing the counter
		v.counter++
	}

	// Validate if the URL is reach the threshold
	isReachCounter := r.isCounterReachThreshold(*v)

	return v.counter, !isReachCounter, nil
}

func (r *rateLimiter) elementsRead(key uuid.UUID) (*value, bool) {
	r.mut.RLock()
	defer r.mut.RUnlock()
	v, ok := r.elements[key]

	return v, ok
}

func (r *rateLimiter) elementsStore(key uuid.UUID, v *value) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.elements[key] = v
}

func (r *rateLimiter) elementsDelete(key uuid.UUID) {
	r.mut.Lock()
	defer r.mut.Unlock()
	delete(r.elements, key)
}

func (r *rateLimiter) Stop() {
	r.ctx.Done()
}

func (r *rateLimiter) elementLock(key uuid.UUID) func() {
	mutValue, _ := r.mutexes.LoadOrStore(key, &sync.Mutex{})
	mut := mutValue.(*sync.Mutex)
	mut.Lock()
	return func() {
		mut.Unlock()
	}
}

// isReachTtl return true in case the TTL has passed
func (r *rateLimiter) isReachTtl(value value, utcTime time.Time) bool {
	return utcTime.After(value.ttl)
}

// isCounterReachThreshold return true in case the elements reach the threshold
func (r *rateLimiter) isCounterReachThreshold(value value) bool {
	return value.counter > r.threshold
}

// clean removing old elements from the elements map. It makes sure to clean unused URLs
func (r *rateLimiter) clean() {
	log.Printf("Staring clean")
	utcTime := time.Now().UTC()
	for key, v := range r.elements {
		if r.isReachTtl(*v, utcTime) {
			unlock := r.elementLock(key)

			// Do double check in case something was change till we got a elementLock for the element
			v2, _ := r.elementsRead(key)
			if r.isReachTtl(*v2, utcTime) {
				// Delete from elements
				r.elementsDelete(key)
				// Delete from mutexes too
				r.mutexes.Delete(key)
				log.Printf("Key '%s' was released", key)
			}

			unlock()
		}
	}
	log.Printf("Clean was successfully finished")
}
