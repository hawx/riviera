package river

import (
	"net/http"
	"sync"
	"time"

	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

// confluence manages a list of streams and aggregates the latest updates into a
// single (truncated) list.
type confluence struct {
	store   persistence.River
	mu      sync.Mutex
	streams map[string]Tributary
	evs     *events
}

func newConfluence(store persistence.River, evs *events) *confluence {
	return &confluence{
		store:   store,
		streams: map[string]Tributary{},
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
	name := stream.Name()
	c.mu.Lock()

	if _, exists := c.streams[name]; exists {
		return
	}

	c.streams[name] = stream
	c.mu.Unlock()

	go func(stream Tributary, name string) {
		feeds := stream.Feeds()
		fetches := stream.Fetches()

		for {
			select {
			case feed := <-feeds:
				c.store.Add(feed)

			case code := <-fetches:
				if code == http.StatusGone {
					c.Remove(name)
				}

				c.evs.Prepend(Event{
					At:   time.Now().UTC(),
					Uri:  stream.Name(),
					Code: code,
				})
			}
		}
	}(stream, name)
}

func (c *confluence) Remove(uri string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if stream, exists := c.streams[uri]; exists {
		stream.Stop()
		delete(c.streams, uri)
		return true
	}

	return false
}
