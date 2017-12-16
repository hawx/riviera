// Package data provides an interface for saving and retrieving data from a
// key-value database arranged into buckets.
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
