package persistence

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/riverjs"
)

func TestRiver(t *testing.T) {
	assert := assert.New(t)
	db := memdata.Open()

	riv, err := NewRiver(db, -time.Minute)
	assert.Nil(err)

	now := time.Now().Round(time.Second)
	feeds := []riverjs.Feed{
		{FeedTitle: "cool", FeedUrl: "http://cool", WhenLastUpdate: riverjs.RssTime{now}},
		{FeedTitle: "what", FeedUrl: "http://what", WhenLastUpdate: riverjs.RssTime{now}},
		{FeedTitle: "hey", FeedUrl: "http://hey", WhenLastUpdate: riverjs.RssTime{now}},
		{FeedTitle: "hey2", FeedUrl: "http://hey", WhenLastUpdate: riverjs.RssTime{now.Add(-10 * time.Second)}},
		{FeedTitle: "hey", FeedUrl: "http://hey", WhenLastUpdate: riverjs.RssTime{now.Add(-2 * time.Second)}},
	}
	for _, feed := range feeds {
		riv.Add(feed)
	}

	// old feed, ignored
	oldfeed := riverjs.Feed{FeedTitle: "out", FeedUrl: "out", WhenLastUpdate: riverjs.RssTime{time.Now().Add(-2 * time.Minute)}}
	riv.Add(oldfeed)

	latest := riv.Latest()
	if assert.Len(latest, len(feeds)) {
		// ordered by date, then reverse alphabetically on FeedUrl
		assert.Equal(feeds[1], latest[0])
		assert.Equal(feeds[2], latest[1])
		assert.Equal(feeds[0], latest[2])
		assert.Equal(feeds[4], latest[3])
		assert.Equal(feeds[3], latest[4])
	}

	intriv, _ := riv.(*river)
	intriv.truncate()

	// make sure old feed has been deleted
	intriv.View(func(tx data.ReadTx) error {
		for _, v := range tx.All() {
			var feed riverjs.Feed
			json.Unmarshal(v, &feed)
			assert.NotEqual(oldfeed.FeedTitle, feed.FeedTitle)
		}
		return nil
	})
}
