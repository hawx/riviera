// Package memdata implements data over a set of in memory maps.
package memdata

import (
	"sort"
	"sync"

	"hawx.me/code/riviera/river/data"
)

type database struct {
	buckets map[string]data.Bucket
	mu      sync.Mutex
}

// Open a new in memory database.
func Open() data.Database {
	return &database{buckets: map[string]data.Bucket{}}
}

func (db *database) Bucket(name []byte) (data.Bucket, error) {
	db.mu.Lock()

	if b, ok := db.buckets[string(name)]; ok {
		return b, nil
	}

	b := &bucket{kv: map[string][]byte{}}
	db.buckets[string(name)] = b
	db.mu.Unlock()

	return b, nil
}

func (db *database) Close() error {
	return nil
}

type bucket struct {
	kv map[string][]byte
	mu sync.RWMutex
}

func (b *bucket) View(t func(data.ReadTx) error) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return t(tx{b})
}

func (b *bucket) Update(t func(data.Tx) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return t(tx{b})
}

type tx struct {
	b *bucket
}

func (x tx) Get(key []byte) []byte {
	return x.b.kv[string(key)]
}

func (x tx) Put(key, value []byte) error {
	x.b.kv[string(key)] = value
	return nil
}

func (x tx) Delete(key []byte) error {
	delete(x.b.kv, string(key))
	return nil
}

func (x tx) After(start []byte) [][]byte {
	var ks []string
	for k := range x.b.kv {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	i := sort.Search(len(ks), func(i int) bool {
		return ks[i] >= string(start)
	})

	ks = ks[i:]

	vs := make([][]byte, len(ks))
	for i, k := range ks {
		vs[i] = x.b.kv[k]
	}

	return vs
}

func (x tx) KeysBefore(last []byte) [][]byte {
	var ks []string
	for k := range x.b.kv {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	i := sort.Search(len(ks), func(i int) bool {
		return ks[i] >= string(last)
	})

	if i < len(ks) {
		ks = ks[:i]
	}

	r := make([][]byte, len(ks))
	for i, k := range ks {
		r[i] = []byte(k)
	}

	return r
}

func (x tx) All() [][]byte {
	var ks []string
	for k := range x.b.kv {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	vs := make([][]byte, len(ks))
	for i, k := range ks {
		vs[i] = x.b.kv[k]
	}

	return vs
}
