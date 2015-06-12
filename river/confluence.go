package river

import (
	"net/http"
	"time"

	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

type trib interface {
	OnUpdate(f func(models.Feed))
	OnStatus(f func(int))
	Uri() string
	Kill()
}

// confluence manages a list of streams and aggregates the latest updates into a
// single (truncated) list.
type confluence struct {
	store     persistence.River
	streams   []trib
	latest    []models.Feed
	cutOff    time.Duration
	metastore *metaStore
}

func newConfluence(store persistence.River, metastore *metaStore, cutOff time.Duration) *confluence {
	return &confluence{
		store:     store,
		streams:   []trib{},
		latest:    store.Latest(cutOff),
		cutOff:    cutOff,
		metastore: metastore,
	}
}

func (c *confluence) Latest() []models.Feed {
	yesterday := time.Now().Add(c.cutOff)
	newLatest := []models.Feed{}

	for _, feed := range c.latest {
		if feed.WhenLastUpdate.After(yesterday) {
			newLatest = append(newLatest, feed)
		}
	}

	c.latest = newLatest
	return c.latest
}

func (c *confluence) Meta() []Metadata {
	return c.metastore.List()
}

func (c *confluence) Add(stream trib) {
	c.streams = append(c.streams, stream)

	stream.OnUpdate(func(feed models.Feed) {
		c.latest = append([]models.Feed{feed}, c.latest...)

		c.metastore.Set(stream.Uri(), feeddata{
			Uri:             stream.Uri(),
			FeedUrl:         feed.FeedUrl,
			WebsiteUrl:      feed.WebsiteUrl,
			FeedTitle:       feed.FeedTitle,
			FeedDescription: feed.FeedDescription,
		})

		c.store.Add(feed)
	})

	stream.OnStatus(func(code int) {
		if code == http.StatusGone {
			c.Remove(stream.Uri())
		}

		c.metastore.Log(stream.Uri(), Event{
			At:   time.Now().UTC(),
			Uri:  stream.Uri(),
			Code: code,
		})
	})
}

func (c *confluence) Remove(uri string) bool {
	streams := []trib{}
	ok := false

	for _, stream := range c.streams {
		if stream.Uri() != uri {
			streams = append(streams, stream)
		} else {
			ok = true
			stream.Kill()
			c.metastore.Delete(uri)
		}
	}

	c.streams = streams
	return ok
}
