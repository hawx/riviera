package database

import (
	"github.com/hawx/riviera/river/models"
	"github.com/hawx/riviera/database"

	"encoding/json"
	"time"
)

type River interface {
	Add(models.Feed)
	Today() []models.Feed
}

type river struct {
	database.Bucket
}

func (d *river) Add(feed models.Feed) {
	d.Update(func(tx database.Tx) error {
		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return tx.Put([]byte(key), value)
	})
}

func (d *river) Today() []models.Feed {
	feeds := []models.Feed{}

	d.View(func(tx database.Tx) error {
		min := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)

		for _, v := range tx.After([]byte(min)) {
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
