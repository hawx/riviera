package boltdata

import (
	"fmt"

	"github.com/boltdb/bolt"
	"hawx.me/code/riviera/data"
)

type database struct {
	db *bolt.DB
}

func Open(path string) (data.Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &database{db}, nil
}

func (this *database) Bucket(name []byte) (data.Bucket, error) {
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
	db   *bolt.DB
	name []byte
}

func (this *bucket) View(t func(data.Tx) error) error {
	return this.db.View(func(x *bolt.Tx) error {
		b := x.Bucket([]byte(this.name))
		return t(tx{b})
	})
}

func (this *bucket) Update(t func(data.Tx) error) error {
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

func (this tx) Delete(key []byte) error {
	return this.b.Delete(key)
}

func (this tx) After(start []byte) [][]byte {
	r := [][]byte{}
	c := this.b.Cursor()

	for k, v := c.Seek(start); k != nil; k, v = c.Next() {
		r = append(r, v)
	}

	return r
}

func (this tx) All() [][]byte {
	r := [][]byte{}
	c := this.b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		r = append(r, v)
	}

	return r
}
