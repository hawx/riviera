package persistence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/models"
)

func TestRiver(t *testing.T) {
	assert := assert.New(t)
	db := memdata.Open()

	river, err := NewRiver(db)
	assert.Nil(err)

	now := time.Now().Round(time.Second)
	feeds := []models.Feed{
		{FeedTitle: "cool", FeedUrl: "http://cool", WhenLastUpdate: models.RssTime{now}},
		{FeedTitle: "what", FeedUrl: "http://what", WhenLastUpdate: models.RssTime{now}},
		{FeedTitle: "hey", FeedUrl: "http://hey", WhenLastUpdate: models.RssTime{now}},
		{FeedTitle: "hey2", FeedUrl: "http://hey", WhenLastUpdate: models.RssTime{now.Add(-10 * time.Second)}},
		{FeedTitle: "hey", FeedUrl: "http://hey", WhenLastUpdate: models.RssTime{now.Add(-2 * time.Second)}},
	}
	for _, feed := range feeds {
		river.Add(feed)
	}

	// old feed, ignored
	river.Add(models.Feed{FeedTitle: "out", FeedUrl: "out", WhenLastUpdate: models.RssTime{time.Now().Add(-2 * time.Minute)}})

	latest := river.Latest(-time.Minute)
	if assert.Len(latest, len(feeds)) {
		// ordered by date, then reverse alphabetically on FeedUrl
		assert.Equal(feeds[1], latest[0])
		assert.Equal(feeds[2], latest[1])
		assert.Equal(feeds[0], latest[2])
		assert.Equal(feeds[4], latest[3])
		assert.Equal(feeds[3], latest[4])
	}
}
