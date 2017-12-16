package memdata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/riverjs"
)

func TestPersistedRiver(t *testing.T) {
	assert := assert.New(t)

	db := Open()

	riv, err := db.Confluence()
	assert.Nil(err)

	now := time.Now().Round(time.Second)
	feeds := []riverjs.Feed{
		{FeedTitle: "cool", FeedUrl: "http://cool", WhenLastUpdate: riverjs.Time(now)},
		{FeedTitle: "what", FeedUrl: "http://what", WhenLastUpdate: riverjs.Time(now)},
		{FeedTitle: "hey", FeedUrl: "http://hey", WhenLastUpdate: riverjs.Time(now)},
		{FeedTitle: "hey2", FeedUrl: "http://hey", WhenLastUpdate: riverjs.Time(now.Add(-10 * time.Second))},
		{FeedTitle: "hey", FeedUrl: "http://hey", WhenLastUpdate: riverjs.Time(now.Add(-2 * time.Second))},
	}
	for _, feed := range feeds {
		riv.Add(feed)
	}

	// old feed, ignored
	oldfeed := riverjs.Feed{FeedTitle: "out", FeedUrl: "out", WhenLastUpdate: riverjs.Time(time.Now().Add(-2 * time.Minute))}
	riv.Add(oldfeed)

	latest := riv.Latest(-time.Minute)
	if assert.Len(latest, len(feeds)) {
		// ordered by date, then reverse alphabetically on FeedUrl
		assert.Equal(feeds[1], latest[0])
		assert.Equal(feeds[2], latest[1])
		assert.Equal(feeds[0], latest[2])
		assert.Equal(feeds[4], latest[3])
		assert.Equal(feeds[3], latest[4])
	}

	riv.Truncate(-time.Minute)

	assert.Len(riv.(*confluenceDatabase).feeds, len(feeds))
}