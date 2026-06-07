package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	tokens float64
	rate   float64
	burst  float64
	last   time.Time
}

func NewLimiter(rate, burst float64) *Limiter {
	return &Limiter{
		tokens: burst,
		rate:   rate,
		burst:  burst,
		last:   time.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	delta := now.Sub(l.last).Seconds()
	l.tokens += delta * l.rate
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.last = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}
