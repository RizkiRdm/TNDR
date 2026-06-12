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

// NewLimiter creates a token bucket rate limiter.
// rate is tokens added per second, burst is the maximum token bucket size.
func NewLimiter(rate, burst float64) *Limiter {
	return &Limiter{
		tokens: burst,
		rate:   rate,
		burst:  burst,
		last:   time.Now(),
	}
}

// Allow reports whether an action can be performed now.
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
