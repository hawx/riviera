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
	streams []Tributary
	evs     *events
}

func newConfluence(store persistence.River, evs *events) *confluence {
	return &confluence{
		store:   store,
		streams: []Tributary{},
		evs:     evs,
	}
}

func (c *confluence) Latest() []models.Feed {
	return c.store.Latest()
}

func (c *confluence) Log() []Event {
	return c.evs.List()
}

func (c *confluence) Add(stream Tributary) {
	c.streams = append(c.streams, stream)

	go func() {
		feeds := stream.Feeds()
		fetches := stream.Fetches()

		for {
			select {
			case feed := <-feeds:
				c.store.Add(feed)

			case code := <-fetches:
				if code == http.StatusGone {
					c.Remove(stream.Name())
				}

				c.evs.Prepend(Event{
					At:   time.Now().UTC(),
					Uri:  stream.Name(),
					Code: code,
				})
			}
		}
	}()
}

func (c *confluence) Remove(uri string) bool {
	idx := -1

	for i, stream := range c.streams {
		if stream.Name() == uri {
			idx = i
			stream.Stop()
			break
		}
	}

	c.streams = append(c.streams[:idx], c.streams[idx+1:]...)
	return idx > 0
}
