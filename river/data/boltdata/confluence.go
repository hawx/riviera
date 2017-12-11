package boltdata

import (
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
	"hawx.me/code/riviera/river/riverjs"

	"encoding/json"
	"time"
)

type Database interface {
	Add(feed riverjs.Feed)
	Truncate(cutoff time.Duration)
	Latest(cutoff time.Duration) []riverjs.Feed
}

// A confluenceDatabase contains persisted feed data, specifically each "block"
// of updates for a feed. This allows the river to be recreated from past data,
// to be displayed on startup.
type confluenceDatabase struct {
	db *bolt.DB
}

var riverBucketName = []byte("river")

func newConfluenceDatabase(db *bolt.DB) (Database, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(riverBucketName)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("bucket: %s", err)
	}

	return &confluenceDatabase{db}, nil
}

func (d *confluenceDatabase) Add(feed riverjs.Feed) {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(riverBucketName)
		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return b.Put([]byte(key), value)
	})
}

func (d *confluenceDatabase) Truncate(cutoff time.Duration) {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(riverBucketName)
		max := time.Now().UTC().Add(cutoff).Format(time.RFC3339)

		last := []byte(max)
		c := b.Cursor()
		for k, _ := c.First(); bytes.Compare(k, last) < 0; k, _ = c.Next() {
			b.Delete(k)
		}

		return nil
	})
}

func (d *confluenceDatabase) Latest(cutoff time.Duration) []riverjs.Feed {
	feeds := []riverjs.Feed{}

	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(riverBucketName)
		min := time.Now().UTC().Add(cutoff).Format(time.RFC3339)

		start := []byte(min)
		c := b.Cursor()

		for k, v := c.Seek(start); k != nil; k, v = c.Next() {
			var feed riverjs.Feed
			json.Unmarshal(v, &feed)
			feeds = append(feeds, feed)
		}

		return nil
	})

	for i := 0; i < len(feeds)/2; i++ {
		feeds[i], feeds[len(feeds)-i-1] = feeds[len(feeds)-i-1], feeds[i]
	}

	return feeds
}
