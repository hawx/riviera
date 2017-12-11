package confluence

import (
	"hawx.me/code/riviera/river/data"
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
	data.Bucket
}

var riverBucketName = []byte("river")

func newConfluenceDatabase(database data.Database) (Database, error) {
	b, err := database.Bucket(riverBucketName)
	if err != nil {
		return nil, err
	}

	return &confluenceDatabase{b}, nil
}

func (d *confluenceDatabase) Add(feed riverjs.Feed) {
	d.Update(func(tx data.Tx) error {
		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return tx.Put([]byte(key), value)
	})
}

func (d *confluenceDatabase) Truncate(cutoff time.Duration) {
	d.Update(func(tx data.Tx) error {
		max := time.Now().UTC().Add(cutoff).Format(time.RFC3339)

		for _, k := range tx.KeysBefore([]byte(max)) {
			tx.Delete(k)
		}

		return nil
	})
}

func (d *confluenceDatabase) Latest(cutoff time.Duration) []riverjs.Feed {
	feeds := []riverjs.Feed{}

	d.View(func(tx data.ReadTx) error {
		min := time.Now().UTC().Add(cutoff).Format(time.RFC3339)

		for _, v := range tx.After([]byte(min)) {
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
