package confluence

import (
	"log"

	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/riverjs"

	"encoding/json"
	"time"
)

// A confluenceDatabase contains persisted feed data, specifically each "block" of updates
// for a feed. This allows the river to be recreated from past data, to be
// displayed on startup.
type confluenceDatabase struct {
	data.Bucket
	cutoff time.Duration
}

var riverBucketName = []byte("river")

func newConfluenceDatabase(database data.Database, cutoff time.Duration) (*confluenceDatabase, error) {
	b, err := database.Bucket(riverBucketName)
	if err != nil {
		return nil, err
	}

	riv := &confluenceDatabase{b, cutoff}

	go func() {
		for _ = range time.Tick(cutoff) {
			log.Println("truncating feed data")
			riv.truncate()
			log.Println("done truncating")
		}
	}()

	return riv, nil
}

func (d *confluenceDatabase) Add(feed riverjs.Feed) {
	d.Update(func(tx data.Tx) error {
		key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedUrl
		value, _ := json.Marshal(feed)

		return tx.Put([]byte(key), value)
	})
}

func (d *confluenceDatabase) truncate() {
	d.Update(func(tx data.Tx) error {
		max := time.Now().UTC().Add(d.cutoff).Format(time.RFC3339)

		for _, k := range tx.KeysBefore([]byte(max)) {
			tx.Delete(k)
		}

		return nil
	})
}

func (d *confluenceDatabase) Latest() []riverjs.Feed {
	feeds := []riverjs.Feed{}

	d.View(func(tx data.ReadTx) error {
		min := time.Now().UTC().Add(d.cutoff).Format(time.RFC3339)

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
