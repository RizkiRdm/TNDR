package cache

import (
	"go.etcd.io/bbolt"
	"time"
)

type Disk struct {
	db *bbolt.DB
}

func NewDisk(path string) (*Disk, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("cache"))
		return err
	})
	return &Disk{db: db}, err
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
	return string(val), true
}

func (d *Disk) Set(key, value string) {
	d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		if b == nil {
			return nil
		}
		return b.Put([]byte(key), []byte(value))
	})
}
