package memdata

import (
	"sync"

	"hawx.me/code/riviera/feed"
)

type feedDatabase struct {
	known map[string]struct{}
	sync.RWMutex
}

// NewDatabase returns an empty in-memory database for item keys.
func newFeedDatabase() (feed.Database, error) {
	return &feedDatabase{known: map[string]struct{}{}}, nil
}

// Contains checks the database for the Key of a feed item and returns true if
// the item has been seen before.
func (d *feedDatabase) Contains(key string) bool {
	d.RLock()
	_, ok := d.known[key]
	d.RUnlock()

	if ok {
		return true
	}

	d.Lock()
	defer d.Unlock()

	d.known[key] = struct{}{}
	return false
}
