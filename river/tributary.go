package river

import (
	"log"
	"net/http"
	"time"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

type tributary struct {
	uri      string
	feed     *feed.Feed
	client   *http.Client
	mapping  Mapping
	onUpdate []func(models.Feed)
	onStatus []func(int)
	quit     chan struct{}
}

func newTributary(store persistence.Bucket, uri string, cacheTimeout time.Duration, mapping Mapping) *tributary {
	p := &tributary{}
	p.uri = uri
	p.feed = feed.New(cacheTimeout, p.itemHandler, store)
	p.client = &http.Client{Timeout: time.Minute, Transport: &statusTransport{http.DefaultTransport.(*http.Transport), p}}
	p.mapping = mapping
	p.quit = make(chan struct{})

	go p.poll()
	return p
}

func (t *tributary) OnUpdate(f func(models.Feed)) {
	t.onUpdate = append(t.onUpdate, f)
}

func (t *tributary) OnStatus(f func(int)) {
	t.onStatus = append(t.onStatus, f)
}

func (t *tributary) Uri() string {
	return t.uri
}

func (t *tributary) poll() {
	log.Println("started fetching", t.uri)
	t.fetch()

loop:
	for {
		select {
		case <-t.quit:
			break loop
		case <-time.After(t.feed.DurationTillUpdate()):
			log.Println("fetching", t.uri)
			t.fetch()
		}
	}

	log.Println("stopped fetching", t.uri)
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
		log.Println(t.trib.uri, "moved to", newLoc)
		t.trib.uri = newLoc
	}

	return
}

// fetch retrieves the feed for the tributary.
func (t *tributary) fetch() {
	code, err := t.feed.Fetch(t.uri, t.client, charset.NewReader)
	t.status(code)

	if err != nil {
		log.Println("error fetching", t.uri+":", code, err)
		return
	}
}

func (t *tributary) itemHandler(feed *feed.Feed, ch *feed.Channel, newitems []*feed.Item) {
	items := []models.Item{}
	for _, item := range newitems {
		converted := t.mapping(item)

		if converted != nil {
			items = append(items, *converted)
		}
	}

	log.Println(len(items), "new item(s) in", t.uri)
	if len(items) == 0 {
		return
	}

	feedUrl := t.uri
	websiteUrl := ""
	for _, link := range ch.Links {
		if feedUrl != "" && websiteUrl != "" {
			break
		}

		if link.Rel == "self" {
			feedUrl = link.Href
		} else {
			websiteUrl = link.Href
		}
	}

	t.notify(models.Feed{
		FeedUrl:         feedUrl,
		WebsiteUrl:      websiteUrl,
		FeedTitle:       ch.Title,
		FeedDescription: ch.Description,
		WhenLastUpdate:  models.RssTime{time.Now()},
		Items:           items,
	})
}

func (t *tributary) notify(feed models.Feed) {
	for _, f := range t.onUpdate {
		f(feed)
	}
}

func (t *tributary) status(code int) {
	for _, f := range t.onStatus {
		f(code)
	}
}

func (t *tributary) Kill() {
	t.onUpdate = []func(models.Feed){}
	close(t.quit)
}
