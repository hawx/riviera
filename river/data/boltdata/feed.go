package boltdata

import (
	"fmt"

	"github.com/boltdb/bolt"
	"hawx.me/code/riviera/feed"
)

// A bucket is an implementation of a feed.Database that persists data.
type feedDatabase struct {
	db   *bolt.DB
	name []byte
}

var in = []byte("in")

func newFeedDatabase(db *bolt.DB, name string) (feed.Database, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("bucket: %s", err)
	}

	return &feedDatabase{db: db, name: []byte(name)}, nil
}

func (d *feedDatabase) Contains(key string) bool {
	ok := false

	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.name)
		if b.Get([]byte(key)) != nil {
			ok = true
		}
		return nil
	})

	if ok {
		return true
	}

	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.name)
		return b.Put([]byte(key), in)
	})

	return false
}
