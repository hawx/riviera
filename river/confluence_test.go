package river

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/internal/persistence"
)

func TestConfluence(t *testing.T) {
	db := memdata.Open()
	river, _ := persistence.NewRiver(db)

	c := newConfluence(river, newEvents(3), time.Minute)

	assert.Empty(t, c.Latest())
}

// type dummyTrib struct {
// 	feed   models.Feed
// 	killed bool
// 	f      func(models.Feed)
// }

// func (d *dummyTrib) push() {
// 	d.f(d.feed)
// }

// func (d *dummyTrib) OnUpdate(f func(models.Feed)) {
// 	d.f = f
// 	d.push()
// }

// func (d *dummyTrib) OnStatus(f func(int)) {

// }

// func (d *dummyTrib) Uri() string { return "hey" }
// func (d *dummyTrib) Kill() {
// 	d.f = func(models.Feed) {}
// 	d.killed = true
// }

// func TestConfluenceWithTributary(t *testing.T) {
// 	db := memdata.Open()
// 	river, _ := persistence.NewRiver(db)

// 	c := newConfluence(river, newEvents(3), -time.Minute)

// 	feed := models.Feed{
// 		FeedTitle:      "hey",
// 		WhenLastUpdate: models.RssTime{time.Now().Add(-time.Second)},
// 	}
// 	feed2 := models.Feed{
// 		FeedTitle:      "cool",
// 		WhenLastUpdate: models.RssTime{time.Now().Add(-5 * time.Second)},
// 	}

// 	trib := &dummyTrib{feed: feed}
// 	trib2 := &dummyTrib{feed: feed2}
// 	c.Add(trib)
// 	assert.Equal(t, []models.Feed{feed}, c.Latest())
// 	c.Add(trib2)
// 	assert.Equal(t, []models.Feed{feed2, feed}, c.Latest())

// 	c.Remove(trib.Uri())
// 	assert.True(t, trib.killed)

// 	trib.push()
// 	assert.Equal(t, []models.Feed{feed2, feed}, c.Latest())

// 	c.Add(trib)
// 	assert.Equal(t, []models.Feed{feed, feed2, feed}, c.Latest())

// 	trib.push()
// 	assert.Equal(t, []models.Feed{feed, feed, feed2, feed}, c.Latest())
// }

// func TestConfluenceWithTributaryWhenTooOld(t *testing.T) {
// 	db := memdata.Open()
// 	river, _ := persistence.NewRiver(db)

// 	c := newConfluence(river, newMetaStore(3), -time.Minute)

// 	feed := models.Feed{
// 		FeedTitle:      "hey",
// 		WhenLastUpdate: models.RssTime{time.Now().Add(-2 * time.Minute)},
// 	}

// 	trib := &dummyTrib{feed: feed}
// 	c.Add(trib)
// 	assert.Empty(t, c.Latest())
// }
