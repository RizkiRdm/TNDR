package cache

import (
	"testing"
	"time"
)

func TestDiskCache_SetAndGet(t *testing.T) {
	tmpFile := t.TempDir() + "/cache.db"
	d, err := NewDisk(tmpFile, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create disk cache: %v", err)
	}
	defer d.db.Close()

	d.Set("key1", "value1")
	val, ok := d.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("expected value1, got %s (ok: %v)", val, ok)
	}
}

func TestDiskCache_MissOnUnknownKey(t *testing.T) {
	tmpFile := t.TempDir() + "/cache.db"
	d, err := NewDisk(tmpFile, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create disk cache: %v", err)
	}
	defer d.db.Close()

	val, ok := d.Get("nonexistent")
	if ok || val != "" {
		t.Errorf("expected empty value and ok=false, got %s (ok: %v)", val, ok)
	}
}

func TestDiskCache_TTLExpiry(t *testing.T) {
	tmpFile := t.TempDir() + "/cache.db"
	ttl := 100 * time.Millisecond
	d, err := NewDisk(tmpFile, ttl)
	if err != nil {
		t.Fatalf("failed to create disk cache: %v", err)
	}
	defer d.db.Close()

	d.Set("key", "val")
	time.Sleep(150 * time.Millisecond)

	val, ok := d.Get("key")
	if ok || val != "" {
		t.Errorf("expected empty value and ok=false after TTL, got %s (ok: %v)", val, ok)
	}
}

func TestDiskCache_PersistsAcrossReopen(t *testing.T) {
	tmpFile := t.TempDir() + "/cache.db"

	d1, err := NewDisk(tmpFile, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create disk cache: %v", err)
	}
	d1.Set("key", "val")
	d1.db.Close()

	d2, err := NewDisk(tmpFile, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to reopen disk cache: %v", err)
	}
	defer d2.db.Close()

	val, ok := d2.Get("key")
	if !ok || val != "val" {
		t.Errorf("expected val, got %s (ok: %v)", val, ok)
	}
}
