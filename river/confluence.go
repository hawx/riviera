package river

import (
	"github.com/hawx/riviera/river/persistence"
	"github.com/hawx/riviera/river/models"

	"time"
)

type confluence struct {
	store   persistence.River
	streams []*tributary
	latest  []models.Feed
	cutOff  time.Duration
}

func newConfluence(store persistence.River, cutOff time.Duration) *confluence {
	return &confluence{store, []*tributary{}, store.Latest(cutOff), cutOff}
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

func (c *confluence) Add(stream *tributary) {
	c.streams = append(c.streams, stream)

	stream.OnUpdate(func(feed models.Feed) {
		c.latest = append([]models.Feed{feed}, c.latest...)
		c.store.Add(feed)
	})
}

func (c *confluence) Remove(uri string) bool {
	streams := []*tributary{}
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
