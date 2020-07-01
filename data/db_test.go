package data

import (
	"testing"

	"hawx.me/code/assert"
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
