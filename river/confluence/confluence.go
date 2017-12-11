package confluence

import (
	"net/http"
	"sync"
	"time"

	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/events"
	"hawx.me/code/riviera/river/riverjs"
	"hawx.me/code/riviera/river/tributary"
)

type Confluence interface {
	Latest() []riverjs.Feed
	Log() []events.Event
	Add(stream tributary.Tributary)
	Remove(uri string) bool
	Close() error
}

// confluence manages a list of streams and aggregates the latest updates into a
// single (truncated) list.
type confluence struct {
	store   *confluenceDatabase
	mu      sync.Mutex
	streams map[string]tributary.Tributary
	feeds   chan riverjs.Feed
	events  chan events.Event
	evs     *events.Events
	quit    chan struct{}
}

func New(store data.Database, cutoff time.Duration, evs *events.Events) Confluence {
	database, _ := newConfluenceDatabase(store, cutoff)

	c := &confluence{
		store:   database,
		streams: map[string]tributary.Tributary{},
		feeds:   make(chan riverjs.Feed),
		events:  make(chan events.Event),
		evs:     evs,
		quit:    make(chan struct{}),
	}

	go c.run()
	return c
}

func (c *confluence) Latest() []riverjs.Feed {
	return c.store.Latest()
}

func (c *confluence) Log() []events.Event {
	return c.evs.List()
}

func (c *confluence) Add(stream tributary.Tributary) {
	name := stream.Name()
	c.mu.Lock()

	if _, exists := c.streams[name]; exists {
		return
	}

	c.streams[name] = stream
	c.mu.Unlock()

	stream.Feeds(c.feeds)
	stream.Events(c.events)
}

func (c *confluence) run() {
loop:
	for {
		select {
		case feed := <-c.feeds:
			c.store.Add(feed)

		case event := <-c.events:
			if event.Code == http.StatusGone {
				c.Remove(event.Uri)
			}
			c.evs.Prepend(event)

		case <-c.quit:
			for _, trib := range c.streams {
				trib.Stop()
			}
			break loop
		}
	}

	close(c.quit)
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

func (c *confluence) Close() error {
	c.quit <- struct{}{}
	<-c.quit

	return nil
}
