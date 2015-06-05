package subscriptions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/data/memdata"
)

func TestSubscriptions(t *testing.T) {
	db := memdata.Open()

	subs, err := Open(db)

	assert.Nil(t, err)
	assert.Equal(t, []Subscription{}, subs.List())

	// Add feed
	subs.Add("http://example.com/feed")
	subs.Add("http://example.org/xml")
	subs.Add("http://example.com/feed2")

	assert.Equal(t, []Subscription{
		{Uri: "http://example.com/feed"},
		{Uri: "http://example.com/feed2"},
		{Uri: "http://example.org/xml"},
	}, subs.List())

	// Refresh feed
	subs.Refresh(Subscription{
		Uri:    "http://example.com/feed",
		Status: Bad,
	})
	assert.Equal(t, []Subscription{
		{Uri: "http://example.com/feed", Status: Bad},
		{Uri: "http://example.com/feed2"},
		{Uri: "http://example.org/xml"},
	}, subs.List())

	// Remove feed
	subs.Remove("http://example.com/feed")
	assert.Equal(t, []Subscription{
		{Uri: "http://example.com/feed2"},
		{Uri: "http://example.org/xml"},
	}, subs.List())
}

func TestOnAdd(t *testing.T) {
	db := memdata.Open()

	subs, _ := Open(db)

	ch := make(chan Subscription, 1)
	subs.OnAdd(func(s Subscription) {
		ch <- s
	})

	subs.Add("http://example.com/feed")

	select {
	case s := <-ch:
		assert.Equal(t, Subscription{Uri: "http://example.com/feed"}, s)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestOnRemove(t *testing.T) {
	db := memdata.Open()

	subs, _ := Open(db)
	subs.Add("http://example.com/feed")

	ch := make(chan string, 1)
	subs.OnRemove(func(s string) {
		ch <- s
	})

	subs.Remove("http://example.com/feed")

	select {
	case s := <-ch:
		assert.Equal(t, "http://example.com/feed", s)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
