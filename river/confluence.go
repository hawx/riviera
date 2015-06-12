package river

import (
	"net/http"
	"time"

	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

// confluence manages a list of streams and aggregates the latest updates into a
// single (truncated) list.
type confluence struct {
	store   persistence.River
	streams []*tributary
	latest  []models.Feed
	cutOff  time.Duration
	evs     *events
}

func newConfluence(store persistence.River, evs *events, cutOff time.Duration) *confluence {
	return &confluence{
		store:   store,
		streams: []*tributary{},
		latest:  store.Latest(cutOff),
		cutOff:  cutOff,
		evs:     evs,
	}
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

func (c *confluence) Log() []Event {
	return c.evs.List()
}

func (c *confluence) Add(stream *tributary) {
	c.streams = append(c.streams, stream)

	stream.OnUpdate = func(feed models.Feed) {
		c.latest = append([]models.Feed{feed}, c.latest...)
		c.store.Add(feed)
	}

	stream.OnStatus = func(code int) {
		if code == http.StatusGone {
			c.Remove(stream.Uri())
		}

		c.evs.Prepend(Event{
			At:   time.Now().UTC(),
			Uri:  stream.Uri(),
			Code: code,
		})
	}
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
