package tributary

import (
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/data"
)

// A bucket is an implementation of a feed.Database that persists data.
type feedDatabase struct {
	data.Bucket
}

var in = []byte("in")

func newFeedDatabase(database data.Database, name string) (feed.Database, error) {
	b, err := database.Bucket([]byte(name))
	if err != nil {
		return nil, err
	}

	return &feedDatabase{b}, nil
}

func (d *feedDatabase) Contains(key string) bool {
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
