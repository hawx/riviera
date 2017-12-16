// Package boltdata implements data over a bolt database.
package boltdata

import (
	"github.com/boltdb/bolt"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/confluence"
	"hawx.me/code/riviera/river/data"
)

type database struct {
	db *bolt.DB
}

// Open the boltdb file at the path.
func Open(path string) (data.Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &database{db}, nil
}

func (d *database) Confluence() (confluence.Database, error) {
	return newConfluenceDatabase(d.db)
}

func (d *database) Feed(name string) (feed.Database, error) {
	return newFeedDatabase(d.db, name)
}

func (d *database) Close() error {
	return d.db.Close()
}
