// Package river generates river.js files. See riverjs.org for more information
// on the format.
package river

import (
	"hawx.me/code/riviera/data"
	"hawx.me/code/riviera/river/models"
	"hawx.me/code/riviera/river/persistence"
	"hawx.me/code/riviera/subscriptions"

	"encoding/json"
	"io"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type River interface {
	WriteTo(io.Writer) error
}

type Options struct {
	// Mapping is the function used to convert a feed item to an item in the
	// river.
	Mapping Mapping

	// CutOff is the duration after which items are not shown in the river. This
	// is given as a negative time and is calculated from the time the feed was
	// fetched not the time the item was published.
	CutOff time.Duration

	// Refresh is the minimum refresh period. If an rss feed does not specify
	// when to be fetched this duration will be used.
	Refresh time.Duration
}

var DefaultOptions = Options{
	Mapping: DefaultMapping,
	CutOff:  -24 * time.Hour,
	Refresh: 15 * time.Minute,
}

type river struct {
	confluence   *confluence
	store        data.Database
	cacheTimeout time.Duration
	subs         subscriptions.List
	mapping      Mapping
}

func New(store data.Database, subs subscriptions.List, options Options) River {
	rp, _ := persistence.NewRiver(store)
	confluence := newConfluence(rp, options.CutOff)

	r := &river{confluence, store, options.Refresh, subs, options.Mapping}

	for _, sub := range subs.List() {
		r.Add(sub)
	}

	subs.OnAdd(r.Add)
	subs.OnRemove(r.Remove)

	return r
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

	return json.NewEncoder(w).Encode(models.River{
		Metadata:     metadata,
		UpdatedFeeds: updatedFeeds,
	})
}

func (r *river) Add(sub subscriptions.Subscription) {
	b, _ := persistence.NewBucket(r.store, sub.Uri)

	tributary := newTributary(b, sub.Uri, r.cacheTimeout, r.mapping)

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

func (r *river) Remove(uri string) {
	r.confluence.Remove(uri)
}
