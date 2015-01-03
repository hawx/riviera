package database

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
