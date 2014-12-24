package feed

import "sync"

type Database interface {
	Contains(string) bool
}

type database struct {
	known map[string]struct{}
	sync.RWMutex
}

func NewDatabase() Database {
	return &database{known: map[string]struct{}{}}
}

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
