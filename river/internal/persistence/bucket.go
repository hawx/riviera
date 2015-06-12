package persistence

import (
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/data"
)

// A Bucket is an implementation of a feed.Database that persists data.
type Bucket interface {
	feed.Database
}

type bucket struct {
	data.Bucket
}

var in = []byte("in")

func NewBucket(database data.Database, name string) (Bucket, error) {
	b, err := database.Bucket([]byte(name))
	if err != nil {
		return nil, err
	}

	return &bucket{b}, nil
}

func (d *bucket) Contains(key string) bool {
	ok := false

	d.View(func(tx data.ReadTx) error {
		if tx.Get([]byte(key)) != nil {
			ok = true
		}
		return nil
	})

	if ok {
		return true
	}

	d.Update(func(tx data.Tx) error {
		return tx.Put([]byte(key), in)
	})

	return false
}
