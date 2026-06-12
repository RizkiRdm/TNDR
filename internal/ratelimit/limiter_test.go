package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestLimiter_Allow_WithinBurst(t *testing.T) {
	l := NewLimiter(10, 5)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Errorf("expected Allow() to be true for call %d", i+1)
		}
	}
}

func TestLimiter_Allow_ExceedBurst(t *testing.T) {
	l := NewLimiter(10, 3)
	// Consume burst
	for i := 0; i < 3; i++ {
		l.Allow()
	}
	// 4th call should fail immediately
	if l.Allow() {
		t.Error("expected Allow() to be false after burst exhausted")
	}
}

func TestLimiter_Allow_Refill(t *testing.T) {
	l := NewLimiter(10, 1)
	if !l.Allow() {
		t.Error("expected first Allow() to be true")
	}
	if l.Allow() {
		t.Error("expected second Allow() to be false")
	}
	// Rate is 10 tokens/s, so 1 token takes 100ms.
	// Sleeping for 150ms should guarantee refill.
	time.Sleep(150 * time.Millisecond)
	if !l.Allow() {
		t.Error("expected Allow() to be true after refill")
	}
}

func TestLimiter_Concurrent(t *testing.T) {
	burst := 10
	l := NewLimiter(100, float64(burst))
	var wg sync.WaitGroup
	var count int
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow() {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if count > burst {
		t.Errorf("expected at most %d successes, got %d", burst, count)
	}
}
