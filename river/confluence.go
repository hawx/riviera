package river

import (
	"github.com/hawx/riviera/river/persistence"
	"github.com/hawx/riviera/river/models"

	"time"
)

type Confluence interface {
	Latest() []models.Feed
	Add(Tributary)
	Remove(string) bool
}

type confluence struct {
	store   persistence.River
	streams []Tributary
	latest  []models.Feed
	cutOff  time.Duration
}

func newConfluence(store persistence.River, cutOff time.Duration) Confluence {
	return &confluence{store, []Tributary{}, store.Today(), cutOff}
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

	stream.OnUpdate(func(feed models.Feed) {
		c.latest = append([]models.Feed{feed}, c.latest...)
		c.store.Add(feed)
	})
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
