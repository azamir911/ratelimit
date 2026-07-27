package ratelimit

import "time"

type clock interface {
	now() time.Time
	newTicker(time.Duration) ticker
}

type ticker interface {
	channel() <-chan time.Time
	stop()
}

type realClock struct{}

func (realClock) now() time.Time {
	return time.Now()
}

func (realClock) newTicker(interval time.Duration) ticker {
	return realTicker{Ticker: time.NewTicker(interval)}
}

type realTicker struct {
	*time.Ticker
}

func (t realTicker) channel() <-chan time.Time {
	return t.C
}

func (t realTicker) stop() {
	t.Stop()
}
