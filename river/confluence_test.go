package river

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/events"
	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/riverjs"
)

func TestConfluence(t *testing.T) {
	db := memdata.Open()
	river, _ := persistence.NewRiver(db, -time.Minute)

	c := newConfluence(river, events.New(3))

	assert.Empty(t, c.Latest())
}

type dummyTrib struct {
	name    string
	feed    riverjs.Feed
	stopped bool
	feeds   chan<- riverjs.Feed
	events  chan<- events.Event
}

func newDummyTrib(feed riverjs.Feed, name string) *dummyTrib {
	return &dummyTrib{
		name:    name,
		feed:    feed,
		stopped: true,
	}
}

func (d *dummyTrib) Name() string { return d.name }

func (d *dummyTrib) push() {
	d.feeds <- d.feed
}

func (d *dummyTrib) Feeds(feeds chan<- riverjs.Feed) {
	d.feeds = feeds
}

func (d *dummyTrib) Events(events chan<- events.Event) {
	d.events = events
}

func (d *dummyTrib) Start() {
	d.stopped = false
	d.push()
	time.Sleep(time.Millisecond)
}

func (d *dummyTrib) Stop() {
	d.stopped = true
}

func TestConfluenceWithTributary(t *testing.T) {
	db := memdata.Open()
	river, _ := persistence.NewRiver(db, -time.Minute)

	c := newConfluence(river, events.New(3))

	now := time.Now().Local().Round(time.Second)

	feed := riverjs.Feed{
		FeedTitle:      "hey",
		WhenLastUpdate: riverjs.RssTime{now.Add(-time.Second)},
	}
	feed2 := riverjs.Feed{
		FeedTitle:      "cool",
		WhenLastUpdate: riverjs.RssTime{now.Add(-5 * time.Second)},
	}

	trib := newDummyTrib(feed, "dummy1")
	trib2 := newDummyTrib(feed2, "dummy2")

	c.Add(trib)
	trib.Start()
	assert.Equal(t, []riverjs.Feed{feed}, c.Latest())

	c.Add(trib2)
	trib2.Start()
	assert.Equal(t, []riverjs.Feed{feed, feed2}, c.Latest())

	c.Remove(trib.Name())
	assert.True(t, trib.stopped)

	c.Add(trib)
	trib.Start()
	assert.Equal(t, []riverjs.Feed{feed, feed2}, c.Latest())
}

func TestConfluenceWithTributaryWhenTooOld(t *testing.T) {
	db := memdata.Open()
	river, _ := persistence.NewRiver(db, -time.Minute)

	c := newConfluence(river, events.New(3))

	feed := riverjs.Feed{
		FeedTitle:      "hey",
		WhenLastUpdate: riverjs.RssTime{time.Now().Add(-5 * time.Minute)},
	}

	trib := newDummyTrib(feed, "dummy3")
	c.Add(trib)
	trib.Start()

	assert.Empty(t, c.Latest())
}
