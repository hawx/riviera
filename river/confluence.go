package river

import (
	"time"

	"hawx.me/code/riviera/river/models"
	"hawx.me/code/riviera/river/persistence"
)

type trib interface {
	OnUpdate(f func(models.Feed))
	Uri() string
	Kill()
}

type confluence struct {
	store   persistence.River
	streams []trib
	latest  []models.Feed
	cutOff  time.Duration
}

func newConfluence(store persistence.River, cutOff time.Duration) *confluence {
	return &confluence{store, []trib{}, store.Latest(cutOff), cutOff}
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

func (c *confluence) Add(stream trib) {
	c.streams = append(c.streams, stream)

	stream.OnUpdate(func(feed models.Feed) {
		c.latest = append([]models.Feed{feed}, c.latest...)
		c.store.Add(feed)
	})
}

func (c *confluence) Remove(uri string) bool {
	streams := []trib{}
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
