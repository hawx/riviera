// Package data provides the ability to rebuild previous feeds and remove
// duplicate items.
package data

import (
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/confluence"
)

// Database is a key-value store with data arranged in buckets.
type Database interface {
	// Feed returns a database for storing known items from a named feed.
	Feed(name string) (feed.Database, error)

	// Confluence returns a database for storing past rivers.
	Confluence() (confluence.Database, error)

	// Close releases all database resources.
	Close() error
}
