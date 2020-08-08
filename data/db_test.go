package data

import (
	"testing"
	"time"

	"hawx.me/code/assert"
	"hawx.me/code/riviera/river/riverjs"
)

func TestFeedDB(t *testing.T) {
	assert := assert.Wrap(t)

	db, err := Open("file:TestFeedDB?cache=shared&mode=memory")
	assert(err).Must.Nil()
	defer db.Close()

	feedDB, err := db.Feed("what")
	assert(err).Must.Nil()

	ok := feedDB.Contains("hey")
	if assert(ok).False() {
		ok = feedDB.Contains("hey")
		assert(ok).True()
	}
}

func TestBucket(t *testing.T) {
	assert := assert.Wrap(t)

	db, err := Open("file:TestBucket?cache=shared&mode=memory")
	assert(err).Must.Nil()
	defer db.Close()

	bucket, err := db.Feed("test")
	assert(err).Nil()

	key := "1"
	assert(bucket.Contains(key)).False()
	assert(bucket.Contains(key)).True()

	bucket2, err := db.Feed("test2")
	assert(err).Nil()

	assert(bucket2.Contains(key)).False()
	assert(bucket2.Contains(key)).True()
}

func TestPersistedRiver(t *testing.T) {
	assert := assert.Wrap(t)

	db, err := Open("file:TestPersistedRiver?cache=shared&mode=memory")
	assert(err).Must.Nil()
	defer db.Close()

	riv := db.Confluence()

	now := time.Now().Round(time.Second)
	feeds := []riverjs.Feed{
		{FeedTitle: "cool", FeedURL: "http://cool", WhenLastUpdate: riverjs.Time(now)},
		{FeedTitle: "what", FeedURL: "http://what", WhenLastUpdate: riverjs.Time(now)},
		{FeedTitle: "hey", FeedURL: "http://hey", WhenLastUpdate: riverjs.Time(now)},
		{FeedTitle: "hey2", FeedURL: "http://hey", WhenLastUpdate: riverjs.Time(now.Add(-10 * time.Second))},
		{FeedTitle: "hey", FeedURL: "http://hey", WhenLastUpdate: riverjs.Time(now.Add(-2 * time.Second))},
	}
	for _, feed := range feeds {
		riv.Add(feed)
	}

	// old feed, ignored
	oldfeed := riverjs.Feed{FeedTitle: "out", FeedURL: "out", WhenLastUpdate: riverjs.Time(time.Now().Add(-2 * time.Minute))}
	riv.Add(oldfeed)

	latest := riv.Latest(-time.Minute)
	if assert(latest).Len(len(feeds)) {
		// ordered by date, then reverse alphabetically on FeedURL
		assert(feeds[1]).Equal(latest[0])
		assert(feeds[2]).Equal(latest[1])
		assert(feeds[0]).Equal(latest[2])
		assert(feeds[4]).Equal(latest[3])
		assert(feeds[3]).Equal(latest[4])
	}

	riv.Truncate(-time.Minute)

	// assert.Len(riv.(*confluenceDatabase).feeds, len(feeds))
}
