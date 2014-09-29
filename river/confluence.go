package river

import (
	"github.com/hawx/riviera/river/database"
	"github.com/hawx/riviera/river/models"

	"time"
)

type Confluence interface {
	Latest() []models.Feed
	Add(Tributary)
	Remove(string) bool
}

type confluence struct {
	store   database.River
	streams []Tributary
	latest  []models.Feed
	cutOff  time.Duration
}

func newConfluence(store database.River, streams []Tributary, cutOff time.Duration) Confluence {
	c := &confluence{store, streams, store.Today(), cutOff}
	for _, r := range streams {
		c.run(r)
	}
	return c
}

func (c *confluence) run(r Tributary) {
	go func(in <-chan models.Feed) {
		for v := range in {
			c.latest = append([]models.Feed{v}, c.latest...)
			c.store.Add(v)
		}
	}(r.Latest())
}

func (c *confluence) Latest() []models.Feed {
	yesterday := time.Now().Add(c.cutOff)
	newLatest := []models.Feed{}

	for _, feed := range c.latest {
		if feed.WhenLastUpdate.After(yesterday) {
			newLatest = append(newLatest, feed)
		}
	}

	c.latest = newLatest
	return c.latest
}

func (c *confluence) Add(stream Tributary) {
	c.streams = append(c.streams, stream)
	c.run(stream)
}

func (c *confluence) Remove(uri string) bool {
	streams := []Tributary{}
	ok := false

	for _, stream := range c.streams {
		if stream.Uri() != uri {
			streams = append(streams, stream)
		} else {
			ok = true
			stream.Kill()
		}
	}

	c.streams = streams
	return ok
}
