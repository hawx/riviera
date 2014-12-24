package database

import (
	"github.com/boltdb/bolt"
	"github.com/hawx/riviera/feed"
)

type Bucket interface {
	feed.Database
}

type bucket struct {
	name string
	db   *bolt.DB
}

var in []byte = []byte("in")

func (d *bucket) Contains(key string) bool {
	ok := false

	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.name))
		if b.Get([]byte(key)) != nil {
			ok = true
		}
		return nil
	})

	if ok {
		return true
	}

	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.name))
		return b.Put([]byte(key), in)
	})

	return false
}
