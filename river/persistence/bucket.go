package persistence

import (
	"github.com/hawx/riviera/data"
	"github.com/hawx/riviera/feed"
)

type Bucket interface {
	feed.Database
}

type bucket struct {
	data.Bucket
}

var in []byte = []byte("in")

func NewBucket(database data.Database, name string) (Bucket, error) {
	b, err := database.Bucket([]byte(name))
	if err != nil {
		return nil, err
	}

	return &bucket{b}, nil
}

func (d *bucket) Contains(key string) bool {
	ok := false

	d.View(func(tx data.Tx) error {
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
