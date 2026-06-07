package cache

import (
	"testing"
	"time"
)

func TestExactCache(t *testing.T) {
	c := NewExact(100 * time.Millisecond)
	key := "test-key"
	val := "test-val"

	c.Set(key, val)

	// Hit
	got, ok := c.Get(key)
	if !ok || got != val {
		t.Errorf("expected %s, got %s", val, got)
	}

	// Expire
	time.Sleep(150 * time.Millisecond)
	_, ok = c.Get(key)
	if ok {
		t.Error("expected cache miss after TTL")
	}
}
