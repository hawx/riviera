package memdata

import (
	"sort"

	"hawx.me/code/riviera/data"
)

type database struct {
	buckets map[string]data.Bucket
}

func Open() data.Database {
	return &database{buckets: map[string]data.Bucket{}}
}

func (db *database) Bucket(name []byte) (data.Bucket, error) {
	if b, ok := db.buckets[string(name)]; ok {
		return b, nil
	}

	b := &bucket{kv: map[string][]byte{}}
	db.buckets[string(name)] = b
	return b, nil
}

func (db *database) Close() error {
	return nil
}

type bucket struct {
	kv map[string][]byte
}

func (b *bucket) View(t func(data.Tx) error) error {
	return t(tx{b})
}

func (b *bucket) Update(t func(data.Tx) error) error {
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

	if i < len(ks) {
		ks = ks[i:]
	}

	vs := make([][]byte, len(ks))
	for i, k := range ks {
		vs[i] = x.b.kv[k]
	}

	return vs
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
