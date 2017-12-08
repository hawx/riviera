package feed

import "sync"

// A Database allows the Feed to keep track of items it has already seen before.
type Database interface {
	Contains(string) bool
}

type database struct {
	known map[string]struct{}
	sync.RWMutex
}

// NewDatabase returns an empty in-memory database for item keys.
func NewDatabase() Database {
	return &database{known: map[string]struct{}{}}
}

// Contains checks the database for the Key of a feed item and returns true if
// the item has been seen before.
func (d *database) Contains(key string) bool {
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
