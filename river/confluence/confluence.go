package confluence

import (
	"log"
	"net/http"
	"sync"
	"time"

	"hawx.me/code/riviera/river/events"
	"hawx.me/code/riviera/river/riverjs"
	"hawx.me/code/riviera/river/tributary"
)

// A Confluence manages a list of Tributaries and aggregates the latest updates
// into a single (truncated) list.
type Confluence interface {
	// Latest returns the newest items from all managed Tributaries.
	Latest() []riverjs.Feed

	// Log returns the events that have been triggered by the Tributaries.
	Log() []events.Event

	// Add causes the Confluence to aggregate a new Tributary. If a Tributary with
	// the same name is already managed by the Confluence no action will be taken.
	Add(stream tributary.Tributary)

	// Remove will stop the named Tributary and remove it from the list of those
	// managed by the Confluence.
	Remove(uri string) bool

	// Close stops the Confluence and all managed Tributaries.
	Close() error
}

type confluence struct {
	store   Database
	cutoff  time.Duration
	mu      sync.Mutex
	streams map[string]tributary.Tributary
	feeds   chan riverjs.Feed
	events  chan events.Event
	evs     *events.Events
	quit    chan struct{}
}

// New creates a new Confluence writing to the store. The cutoff specifies the
// minimum duration an item should be returned by Latest for, but is not
// guaranteed to be followed exactly (e.g. with a cutoff of 1 hour an item which
// is 2 hours old may be returned by Latest, but an item that is 5 minutes old
// must be returned by Latest). The event log size is set by logLength.
func New(store Database, cutoff time.Duration, logLength int) Confluence {
	go func() {
		for _ = range time.Tick(cutoff) {
			log.Println("truncating feed data")
			store.Truncate(cutoff)
			log.Println("done truncating")
		}
	}()

	evs := events.New(logLength)

	c := &confluence{
		store:   store,
		cutoff:  cutoff,
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
	return c.store.Latest(c.cutoff)
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
				c.Remove(event.URI)
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
