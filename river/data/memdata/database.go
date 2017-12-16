// Package memdata implements data over a set of in memory maps.
package memdata

import (
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/confluence"
	"hawx.me/code/riviera/river/data"
)

type database struct{}

// Open a new in-memory database.
func Open() data.Database {
	return &database{}
}

func (*database) Confluence() (confluence.Database, error) {
	return newConfluenceDatabase()
}

func (*database) Feed(name string) (feed.Database, error) {
	return newFeedDatabase()
}

func (db *database) Close() error {
	return nil
}
