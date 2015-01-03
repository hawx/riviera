// Package data provides an interface for saving and retrieving data from a
// key-value database arranged into buckets. It uses boltdb as its persistence
// layer.
package data

import (
	"github.com/boltdb/bolt"
	"fmt"
)


type Database interface {
	// Bucket returns a namespaced bucket for storing key-value data.
	Bucket(name []byte) (Bucket, error)

	// Close releases all database resources. All transactions must be complete
	// before closing.
	Close() error
}

type Bucket interface {
	// View executes the function in the context of a read-only transaction.
	View(func(Tx) error) error

	// Update executes the function in the context of a read-write transaction.
	Update(func(Tx) error) error
}

type Tx interface {
	// Get returns the value associated with a key. Returns a nil value if the key does not exist.
	Get(key []byte) []byte

	// Put sets the value associated with a key.
	Put(key, value []byte) error

	// After returns all values listed after the key given to the last value.
	After(start []byte) [][]byte
}


type database struct {
	db *bolt.DB
}

func Open(path string) (Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &database{db}, nil
}

func (this *database) Bucket(name []byte) (Bucket, error) {
	err := this.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("bucket: %s", err)
	}

	return &bucket{this.db, name}, nil
}

func (this *database) Close() error {
	return this.db.Close()
}


type bucket struct {
	db *bolt.DB
	name []byte
}

func (this *bucket) View(t func(Tx) error) error {
	return this.db.View(func(x *bolt.Tx) error {
		b := x.Bucket([]byte(this.name))
		return t(tx{b})
	})
}

func (this *bucket) Update(t func(Tx) error) error {
	return this.db.Update(func(x *bolt.Tx) error {
		b := x.Bucket([]byte(this.name))
		return t(tx{b})
	})
}


type tx struct {
	b *bolt.Bucket
}

func (this tx) Get(key []byte) []byte {
	return this.b.Get(key)
}

func (this tx) Put(key, value []byte) error {
	return this.b.Put(key, value)
}

func (this tx) After(start []byte) [][]byte {
	r := [][]byte{}
	c := this.b.Cursor()

	for k, v := c.Seek(start); k != nil; k, v = c.Next() {
		r = append(r, v)
	}

	return r
}
