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
	cutOff  time.Duration
	evs     *events
}

func newConfluence(store persistence.River, evs *events, cutOff time.Duration) *confluence {
	return &confluence{
		store:   store,
		streams: []*tributary{},
		cutOff:  cutOff,
		evs:     evs,
	}
}

func (c *confluence) Latest() []models.Feed {
	return c.store.Latest(c.cutOff)
}

func (c *confluence) Log() []Event {
	return c.evs.List()
}

func (c *confluence) Add(stream *tributary) {
	c.streams = append(c.streams, stream)

	stream.OnUpdate = func(feed models.Feed) {
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
	idx := -1

	for i, stream := range c.streams {
		if stream.Uri() == uri {
			idx = i
			stream.Kill()
			break
		}
	}

	c.streams = append(c.streams[:idx], c.streams[idx+1:]...)
	return idx > 0
}
