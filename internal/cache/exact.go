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
	mu    sync.RWMutex
	items map[string]Entry
	ttl   time.Duration
}

func NewExact(ttl time.Duration) *Exact {
	return &Exact{
		items: make(map[string]Entry),
		ttl:   ttl,
	}
}

func HashKey(req interface{}) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

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

func (c *Exact) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Entry{
		Value:     value,
		CreatedAt: time.Now(),
	}
}
