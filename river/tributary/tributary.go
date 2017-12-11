package tributary

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/feed/common"
	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/events"
	"hawx.me/code/riviera/river/mapping"
	"hawx.me/code/riviera/river/riverjs"
)

type Tributary interface {
	// Name returns a unique name to refer to the tributary.
	Name() string

	// Feeds sets a channel that is used to send out the latest updates to the
	// tributary.
	Feeds(chan<- riverjs.Feed)

	// Events sets a channel that is used to send out events for the tributary.
	Events(chan<- events.Event)

	// Start polling for updates.
	Start()

	// Stop polling.
	Stop()
}

type tributary struct {
	uri     *url.URL
	feed    *feed.Feed
	client  *http.Client
	mapping mapping.Mapping
	feeds   chan<- riverjs.Feed
	events  chan<- events.Event
	quit    chan struct{}
}

func New(store data.Database, uri string, cacheTimeout time.Duration, mapping mapping.Mapping) *tributary {
	parsedUri, _ := url.Parse(uri)
	feedDatabase, _ := newFeedDatabase(store, uri)

	p := &tributary{
		uri:     parsedUri,
		mapping: mapping,
		quit:    make(chan struct{}),
	}

	p.feed = feed.New(cacheTimeout, p.itemHandler, feedDatabase)
	p.client = &http.Client{Timeout: time.Minute, Transport: &statusTransport{http.DefaultTransport.(*http.Transport), p}}

	return p
}

func (t *tributary) Name() string {
	return t.uri.String()
}

func (t *tributary) Feeds(feeds chan<- riverjs.Feed) {
	t.feeds = feeds
}

func (t *tributary) Events(events chan<- events.Event) {
	t.events = events
}

func (t *tributary) Start() {
	go func() {
		log.Printf("started fetching %s\n", t.uri)
		t.fetch()

	loop:
		for {
			select {
			case <-time.After(t.feed.DurationTillUpdate()):
				log.Printf("fetching %s\n", t.uri)
				t.fetch()
			case <-t.quit:
				break loop
			}
		}

		log.Printf("stopped fetching %s\n", t.uri)
		close(t.quit)
	}()
}

func (t *tributary) Stop() {
	t.quit <- struct{}{}
	<-t.quit
}

type statusTransport struct {
	*http.Transport
	trib *tributary
}

// RoundTrip performs a RoundTrip using the underlying Transport, but then
// checks if the status returned was a 301 MovedPermanently. If so it modifies
// the underlying uri which will then be saved to the subscriptions next time it
// is fetched.
func (t *statusTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.Transport.RoundTrip(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusMovedPermanently {
		newLoc := resp.Header.Get("Location")

		newUrl, errp := url.Parse(newLoc)
		if errp != nil {
			log.Printf("%s moved to %s, error: %s\n", t.trib.uri, newLoc, errp)
			return
		}

		log.Printf("%s moved to %s\n", t.trib.uri, newLoc)
		t.trib.uri = newUrl
	}

	return
}

// fetch retrieves the feed for the tributary.
func (t *tributary) fetch() {
	code, err := t.feed.Fetch(t.uri.String(), t.client, charset.NewReaderLabel)
	t.events <- events.Event{
		At:   time.Now().UTC(),
		Uri:  t.Name(),
		Code: code,
	}

	if err != nil {
		log.Printf("error fetching %s: %d %s\n", t.uri, code, err)
		return
	}
}

func maybeResolvedLink(root *url.URL, other string) string {
	parsed, err := root.Parse(other)
	if err == nil {
		return parsed.String()
	}

	return other
}

func (t *tributary) itemHandler(feed *feed.Feed, ch *common.Channel, newitems []*common.Item) {
	items := []riverjs.Item{}
	for _, item := range newitems {
		converted := t.mapping(item)

		if converted != nil {
			converted.Link = maybeResolvedLink(t.uri, converted.Link)
			converted.PermaLink = maybeResolvedLink(t.uri, converted.PermaLink)

			items = append(items, *converted)
		}
	}

	log.Printf("%d new item(s) in %s\n", len(items), t.uri)
	if len(items) == 0 {
		return
	}

	feedUrl := t.uri.String()
	websiteUrl := ""
	for _, link := range ch.Links {
		if feedUrl != "" && websiteUrl != "" {
			break
		}

		if link.Rel == "self" {
			feedUrl = maybeResolvedLink(t.uri, link.Href)
		} else {
			websiteUrl = maybeResolvedLink(t.uri, link.Href)
		}
	}

	t.feeds <- riverjs.Feed{
		FeedUrl:         feedUrl,
		WebsiteUrl:      websiteUrl,
		FeedTitle:       ch.Title,
		FeedDescription: ch.Description,
		WhenLastUpdate:  riverjs.RssTime{time.Now()},
		Items:           items,
	}
}
