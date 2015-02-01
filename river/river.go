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
	SubscribeTo(subscriptions.List)
}

type river struct {
	confluence   *confluence
	store        data.Database
	cacheTimeout time.Duration
	subs         subscriptions.List
}

func New(store data.Database, cutOff, cacheTimeout time.Duration) River {
	r, _ := persistence.NewRiver(store)
	confluence := newConfluence(r, cutOff)

	return &river{confluence, store, cacheTimeout, nil}
}

func (r *river) SubscribeTo(subs subscriptions.List) {
	r.subs = subs

	for _, sub := range subs.List() {
		r.Add(sub)
	}

	subs.OnAdd(func(sub subscriptions.Subscription) {
		r.Add(sub)
	})

	subs.OnRemove(func(uri string) {
		r.Remove(uri)
	})
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

	tributary.OnStatus(func(code Status) {
		switch code {
		case Good:
			sub.Status = subscriptions.Good
		case Bad:
			sub.Status = subscriptions.Bad
		case Gone:
			sub.Status = subscriptions.Gone
			defer tributary.Kill()
		}

		r.subs.Refresh(sub)
	})

	r.confluence.Add(tributary)
}

func (r *river) Remove(uri string) bool {
	return r.confluence.Remove(uri)
}
