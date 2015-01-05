package persistence

import (
	"github.com/hawx/riviera/river/models"
	"github.com/hawx/riviera/data"

	"encoding/json"
	"time"
)

type River interface {
	Add(models.Feed)
	Latest(time.Duration) []models.Feed
}

type river struct {
	data.Bucket
}

var riverBucketName = []byte("river")

func NewRiver(database data.Database) (River, error) {
	b, err := database.Bucket(riverBucketName)
	if err != nil {
		return nil, err
	}

	return &river{b}, nil
}

func (d *river) Add(feed models.Feed) {
	d.Update(func(tx data.Tx) error {
		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return tx.Put([]byte(key), value)
	})
}

func (d *river) Latest(cutOff time.Duration) []models.Feed {
	feeds := []models.Feed{}

	d.View(func(tx data.Tx) error {
		min := time.Now().UTC().Add(cutOff).Format(time.RFC3339)

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
