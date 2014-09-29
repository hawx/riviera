package database

import (
	"github.com/boltdb/bolt"
	"github.com/hawx/riviera/river/models"

	"encoding/json"
	"time"
)

type River interface {
	Add(models.Feed)
	Today() []models.Feed
}

type river struct {
	db *bolt.DB
}

const riverBucket = "river"

func (d *river) Add(feed models.Feed) {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(riverBucket))

		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return b.Put([]byte(key), value)
	})
}

func (d *river) Today() []models.Feed {
	feeds := []models.Feed{}

	d.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(riverBucket)).Cursor()
		min := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)

		for k, v := c.Seek([]byte(min)); k != nil; k, v = c.Next() {
			var feed models.Feed
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
