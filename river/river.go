// Package river aggregates feeds into a riverjs file.
//
// See http://riverjs.org for more information on the format.
package river

import (
	"encoding/json"
	"io"
	"time"

	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/riverjs"
)

const docsPath = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type River interface {
	// WriteTo writes the river to w in json format. It does not write the json in
	// a callback function.
	WriteTo(w io.Writer) error

	// Log returns a list of fetch events.
	Log() []Event

	// Add subscribes the river to the feed at uri.
	Add(uri string)

	// Remove unsubscribes the river from the feed at url.
	Remove(uri string)

	// Close gracefully stops feeds from being checked.
	Close() error
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

	// LogLength defines the number of events to keep in the crawl log, per feed.
	LogLength int
}

var DefaultOptions = Options{
	Mapping:   DefaultMapping,
	CutOff:    -24 * time.Hour,
	Refresh:   15 * time.Minute,
	LogLength: 0,
}

// river acts as the top-level factory. It manages the creation of the initial
// confluence and creating new tributaries to add to it.
type river struct {
	confluence   *confluence
	store        data.Database
	cacheTimeout time.Duration
	mapping      Mapping
}

// New creates an empty river.
func New(store data.Database, options Options) River {
	if options.Mapping == nil {
		options.Mapping = DefaultOptions.Mapping
	}
	if options.CutOff == 0 {
		options.CutOff = DefaultOptions.CutOff
	}
	if options.Refresh == 0 {
		options.Refresh = DefaultOptions.Refresh
	}

	rp, _ := persistence.NewRiver(store, options.CutOff)

	return &river{
		confluence:   newConfluence(rp, newEvents(options.LogLength)),
		store:        store,
		cacheTimeout: options.Refresh,
		mapping:      options.Mapping,
	}
}

func (r *river) WriteTo(w io.Writer) error {
	updatedFeeds := riverjs.Feeds{r.confluence.Latest()}
	now := time.Now()

	metadata := riverjs.Metadata{
		Docs:      docsPath,
		WhenGMT:   riverjs.RssTime{now.UTC()},
		WhenLocal: riverjs.RssTime{now},
		Version:   "3",
		Secs:      0,
	}

	return json.NewEncoder(w).Encode(riverjs.River{
		Metadata:     metadata,
		UpdatedFeeds: updatedFeeds,
	})
}

func (r *river) Add(uri string) {
	b, _ := persistence.NewBucket(r.store, uri)

	tributary := newTributary(b, uri, r.cacheTimeout, r.mapping)
	r.confluence.Add(tributary)

	tributary.Start()
}

func (r *river) Remove(uri string) {
	r.confluence.Remove(uri)
}

func (r *river) Log() []Event {
	return r.confluence.Log()
}

func (r *river) Close() error {
	r.confluence.Close()
	return nil
}
