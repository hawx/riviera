package river

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

func TestConfluence(t *testing.T) {
	db := memdata.Open()
	river, _ := persistence.NewRiver(db, -time.Minute)

	c := newConfluence(river, newEvents(3))

	assert.Empty(t, c.Latest())
}

type dummyTrib struct {
	name    string
	feed    models.Feed
	stopped bool
	feeds   chan<- models.Feed
	events  chan<- Event
}

func newDummyTrib(feed models.Feed, name string) *dummyTrib {
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

func (d *dummyTrib) Feeds(feeds chan<- models.Feed) {
	d.feeds = feeds
}

func (d *dummyTrib) Events(events chan<- Event) {
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

	c := newConfluence(river, newEvents(3))

	now := time.Now().Local().Round(time.Second)

	feed := models.Feed{
		FeedTitle:      "hey",
		WhenLastUpdate: models.RssTime{now.Add(-time.Second)},
	}
	feed2 := models.Feed{
		FeedTitle:      "cool",
		WhenLastUpdate: models.RssTime{now.Add(-5 * time.Second)},
	}

	trib := newDummyTrib(feed, "dummy1")
	trib2 := newDummyTrib(feed2, "dummy2")

	c.Add(trib)
	trib.Start()
	assert.Equal(t, []models.Feed{feed}, c.Latest())

	c.Add(trib2)
	trib2.Start()
	assert.Equal(t, []models.Feed{feed, feed2}, c.Latest())

	c.Remove(trib.Name())
	assert.True(t, trib.stopped)

	c.Add(trib)
	trib.Start()
	assert.Equal(t, []models.Feed{feed, feed2}, c.Latest())
}

func TestConfluenceWithTributaryWhenTooOld(t *testing.T) {
	db := memdata.Open()
	river, _ := persistence.NewRiver(db, -time.Minute)

	c := newConfluence(river, newEvents(3))

	feed := models.Feed{
		FeedTitle:      "hey",
		WhenLastUpdate: models.RssTime{time.Now().Add(-5 * time.Minute)},
	}

	trib := newDummyTrib(feed, "dummy3")
	c.Add(trib)
	trib.Start()

	assert.Empty(t, c.Latest())
}
