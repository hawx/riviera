package memdata

import (
	"sort"
	"sync"
	"time"

	"hawx.me/code/riviera/river/confluence"
	"hawx.me/code/riviera/river/riverjs"
)

type confluenceDatabase struct {
	mu    sync.RWMutex
	feeds map[string]riverjs.Feed
}

func newConfluenceDatabase() (confluence.Database, error) {
	return &confluenceDatabase{feeds: map[string]riverjs.Feed{}}, nil
}

func (d *confluenceDatabase) Add(feed riverjs.Feed) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := feed.WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feed.FeedURL
	d.feeds[key] = feed
}

func (d *confluenceDatabase) Truncate(cutoff time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	max := time.Now().UTC().Add(cutoff)

	for key, feed := range d.feeds {
		date := feed.WhenLastUpdate.UTC()
		if !max.Before(date) {
			delete(d.feeds, key)
		}
	}
}

func (d *confluenceDatabase) Latest(cutoff time.Duration) []riverjs.Feed {
	d.mu.RLock()
	defer d.mu.RUnlock()

	feeds := []riverjs.Feed{}
	max := time.Now().UTC().Add(cutoff)

	for _, feed := range d.feeds {
		date := feed.WhenLastUpdate.UTC()
		if max.Before(date) {
			feeds = append(feeds, feed)
		}
	}

	sort.Slice(feeds, func(i, j int) bool {
		iKey := feeds[i].WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feeds[i].FeedURL
		jKey := feeds[j].WhenLastUpdate.UTC().Format(time.RFC3339) + " " + feeds[j].FeedURL

		return iKey > jKey
	})

	return feeds
}
