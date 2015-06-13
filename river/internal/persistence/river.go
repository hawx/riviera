package persistence

import (
	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/models"

	"encoding/json"
	"time"
)

// A River contains persisted feed data, specifically each "block" of updates
// for a feed. This allows the river to be recreated from past data, to be
// displayed on startup.
type River interface {
	Add(models.Feed)
	Latest() []models.Feed
}

type river struct {
	data.Bucket
	cutoff time.Duration
}

var riverBucketName = []byte("river")

func NewRiver(database data.Database, cutoff time.Duration) (River, error) {
	b, err := database.Bucket(riverBucketName)
	if err != nil {
		return nil, err
	}

	return &river{b, cutoff}, nil
}

func (d *river) Add(feed models.Feed) {
	d.Update(func(tx data.Tx) error {
		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return tx.Put([]byte(key), value)
	})
}

func (d *river) Latest() []models.Feed {
	feeds := []models.Feed{}

	d.View(func(tx data.ReadTx) error {
		min := time.Now().UTC().Add(d.cutoff).Format(time.RFC3339)

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
