// Package river generates river.js files. See riverjs.org for more information
// on the format.
package river

import (
	"github.com/hawx/riviera/data"
	"github.com/hawx/riviera/subscriptions"
	"github.com/hawx/riviera/river/persistence"
	"github.com/hawx/riviera/river/models"

	"encoding/json"
	"time"
	"io"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type River interface {
	WriteTo(io.Writer) error
}

type river struct {
	confluence   Confluence
	store        data.Database
	cacheTimeout time.Duration
	subs         subscriptions.List
}

func New(store data.Database, subs subscriptions.List, cutOff, cacheTimeout time.Duration) River {
	streams := []Tributary{}

	for _, sub := range subs.List() {
		sub := sub
		bucket, _ := persistence.NewBucket(store, sub.Uri)
		tributary := newTributary(bucket, sub.Uri, cacheTimeout)
		tributary.OnUpdate(func(feed models.Feed) {
			sub.FeedUrl = feed.FeedUrl
			sub.WebsiteUrl = feed.WebsiteUrl
			sub.FeedTitle = feed.FeedTitle
			sub.FeedDescription = feed.FeedDescription

			subs.Refresh(sub)
		})
		streams = append(streams, tributary)
	}

	r, _ := persistence.NewRiver(store)
	confluence := newConfluence(r, streams, cutOff)
	riv := &river{confluence, store, cacheTimeout, subs}

	subs.OnAdd(func(sub subscriptions.Subscription) {
		riv.Add(sub)
	})

	subs.OnRemove(func(uri string) {
		riv.Remove(uri)
	})

	return riv
}

func (r *river) WriteTo(w io.Writer) error {
	updatedFeeds := models.Feeds{r.confluence.Latest()}
	now := time.Now()

	metadata := models.Metadata{
		Docs:      DOCS,
		WhenGMT:   models.RssTime{now.UTC()},
		WhenLocal: models.RssTime{now},
		Version:   "3",
		Secs:      0,
	}

	wrapper := models.Wrapper{
		Metadata:     metadata,
		UpdatedFeeds: updatedFeeds,
	}

	return json.NewEncoder(w).Encode(wrapper)
}

func (r *river) Add(sub subscriptions.Subscription) {
	b, _ := persistence.NewBucket(r.store, sub.Uri)

	tributary := newTributary(b, sub.Uri, r.cacheTimeout)
	tributary.OnUpdate(func(feed models.Feed) {
		sub.FeedUrl = feed.FeedUrl
		sub.WebsiteUrl = feed.WebsiteUrl
		sub.FeedTitle = feed.FeedTitle
		sub.FeedDescription = feed.FeedDescription

		r.subs.Refresh(sub)
	})

	r.confluence.Add(tributary)
}

func (r *river) Remove(uri string) bool {
	return r.confluence.Remove(uri)
}
