package database

import (
	rss "github.com/hawx/go-pkg-rss"
	"github.com/boltdb/bolt"
)

type Bucket interface {
	rss.Database
}

type bucket struct {
	name string
	db   *bolt.DB
}

var in []byte = []byte("in")

func (d *bucket) Get(key string) bool {
	ok := false

	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.name))
		if b.Get([]byte(key)) != nil {
			ok = true
		}
		return nil
	})

	return ok
}

func (d *bucket) Set(key string) {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.name))
		return b.Put([]byte(key), in)
	})
}
