package cache

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"time"
)

type diskEntry struct {
	Value     string    `json:"v"`
	CreatedAt time.Time `json:"t"`
}

type Disk struct {
	db  *bbolt.DB
	ttl time.Duration
}

func NewDisk(path string, ttl time.Duration) (*Disk, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("cache"))
		return err
	})
	return &Disk{db: db, ttl: ttl}, err
}

func (d *Disk) Get(key string) (string, bool) {
	var val []byte
	d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		if b == nil {
			return nil
		}
		val = b.Get([]byte(key))
		return nil
	})
	if len(val) == 0 {
		return "", false
	}

	var entry diskEntry
	if err := json.Unmarshal(val, &entry); err != nil {
		return "", false
	}

	if d.ttl > 0 && time.Since(entry.CreatedAt) > d.ttl {
		d.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("cache"))
			if b != nil {
				b.Delete([]byte(key))
			}
			return nil
		})
		return "", false
	}
	return entry.Value, true
}

func (d *Disk) Set(key, value string) {
	entry := diskEntry{
		Value:     value,
		CreatedAt: time.Now(),
	}
	data, _ := json.Marshal(entry)
	d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		if b == nil {
			return nil
		}
		return b.Put([]byte(key), data)
	})
}
