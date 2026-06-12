package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	Value     string
	CreatedAt time.Time
}

type Exact struct {
	mu       sync.RWMutex
	items    map[string]Entry
	ttl      time.Duration
	maxItems int
}

// NewExact creates a new in-memory exact cache with the given TTL duration.
// Entries older than ttl will not be returned on Get.
func NewExact(ttl time.Duration) *Exact {
	return &Exact{
		items:    make(map[string]Entry),
		ttl:      ttl,
		maxItems: 1000,
	}
}

// HashKey computes a SHA-256 hash of the request to be used as a cache key.
func HashKey(req interface{}) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

// Get retrieves a value from the cache if it exists and hasn't expired.
func (c *Exact) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok {
		return "", false
	}

	if time.Since(entry.CreatedAt) > c.ttl {
		return "", false
	}

	return entry.Value, true
}

// Set adds a value to the cache with the given key.
func (c *Exact) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Kalau sudah penuh, hapus expired entries dulu
	if len(c.items) >= c.maxItems {
		c.evictExpired()
	}

	// Kalau masih penuh setelah evict, hapus random entry
	if len(c.items) >= c.maxItems {
		for k := range c.items {
			delete(c.items, k)
			break
		}
	}

	c.items[key] = Entry{
		Value:     value,
		CreatedAt: time.Now(),
	}
}

// evictExpired harus dipanggil saat lock sudah dipegang
func (c *Exact) evictExpired() {
	for k, v := range c.items {
		if time.Since(v.CreatedAt) > c.ttl {
			delete(c.items, k)
		}
	}
}
