// Package boltdata implements data over a bolt database.
package boltdata

import (
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/confluence"
	"hawx.me/code/riviera/river/data"
)

type database struct {
	db *bolt.DB
}

// Open the boltdb file at the path.
func Open(path string) (data.Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &database{db}, nil
}

func (d *database) Confluence() (confluence.Database, error) {
	return newConfluenceDatabase(d.db)
}

func (d *database) Feed(name string) (feed.Database, error) {
	return newFeedDatabase(d.db, name)
}

func (d *database) Bucket(name []byte) (data.Bucket, error) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("bucket: %s", err)
	}

	return &bucket{d.db, name}, nil
}

func (d *database) Close() error {
	return d.db.Close()
}

type bucket struct {
	db   *bolt.DB
	name []byte
}

func (d *bucket) View(t func(data.ReadTx) error) error {
	return d.db.View(func(x *bolt.Tx) error {
		b := x.Bucket([]byte(d.name))
		return t(tx{b})
	})
}

func (d *bucket) Update(t func(data.Tx) error) error {
	return d.db.Update(func(x *bolt.Tx) error {
		b := x.Bucket([]byte(d.name))
		return t(tx{b})
	})
}

type tx struct {
	b *bolt.Bucket
}

func (t tx) Get(key []byte) []byte {
	return t.b.Get(key)
}

func (t tx) Put(key, value []byte) error {
	return t.b.Put(key, value)
}

func (t tx) Delete(key []byte) error {
	return t.b.Delete(key)
}

func (t tx) After(start []byte) [][]byte {
	r := [][]byte{}
	c := t.b.Cursor()

	for k, v := c.Seek(start); k != nil; k, v = c.Next() {
		r = append(r, v)
	}

	return r
}

func (t tx) KeysBefore(last []byte) [][]byte {
	r := [][]byte{}
	c := t.b.Cursor()

	for k, _ := c.First(); bytes.Compare(k, last) < 0; k, _ = c.Next() {
		r = append(r, k)
	}

	return r
}

func (t tx) All() [][]byte {
	r := [][]byte{}
	c := t.b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		r = append(r, v)
	}

	return r
}
