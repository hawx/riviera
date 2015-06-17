package river

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

type Tributary interface {
	// Name returns a unique name to refer to the tributary.
	Name() string

	// Feeds returns a channel that sends out the latest updates to the tributary.
	Feeds() <-chan models.Feed

	// Fetches returns a channel that sends out the status code for each fetch.
	Fetches() <-chan int

	// Start polling for updates.
	Start()

	// Stop polling.
	Stop()
}

type tributary struct {
	uri     *url.URL
	feed    *feed.Feed
	client  *http.Client
	mapping Mapping
	feeds   chan models.Feed
	fetches chan int
	quit    chan struct{}
}

func newTributary(store persistence.Bucket, uri string, cacheTimeout time.Duration, mapping Mapping) *tributary {
	parsedUri, _ := url.Parse(uri)

	p := &tributary{
		uri:     parsedUri,
		mapping: mapping,
		feeds:   make(chan models.Feed),
		fetches: make(chan int),
		quit:    make(chan struct{}),
	}

	p.feed = feed.New(cacheTimeout, p.itemHandler, store)
	p.client = &http.Client{Timeout: time.Minute, Transport: &statusTransport{http.DefaultTransport.(*http.Transport), p}}

	return p
}

func (t *tributary) Name() string {
	return t.uri.String()
}

func (t *tributary) Feeds() <-chan models.Feed {
	return t.feeds
}

func (t *tributary) Fetches() <-chan int {
	return t.fetches
}

func (t *tributary) Start() {
	go func() {
		log.Printf("started fetching %s\n", t.uri)
		t.fetch()

	loop:
		for {
			select {
			case <-t.quit:
				break loop
			case <-time.After(t.feed.DurationTillUpdate()):
				log.Printf("fetching %s\n", t.uri)
				t.fetch()
			}
		}

		log.Printf("stopped fetching %s\n", t.uri)
	}()
}

func (t *tributary) Stop() {
	close(t.quit)
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
	code, err := t.feed.Fetch(t.uri.String(), t.client, charset.NewReader)
	t.fetches <- code

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

func (t *tributary) itemHandler(feed *feed.Feed, ch *feed.Channel, newitems []*feed.Item) {
	items := []models.Item{}
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

	t.feeds <- models.Feed{
		FeedUrl:         feedUrl,
		WebsiteUrl:      websiteUrl,
		FeedTitle:       ch.Title,
		FeedDescription: ch.Description,
		WhenLastUpdate:  models.RssTime{time.Now()},
		Items:           items,
	}
}
