package river

import (
	"github.com/hawx/riviera/feed"
	"github.com/hawx/riviera/river/models"
	"github.com/hawx/riviera/river/persistence"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"

	"io"
	"log"
	"net/http"
	"time"
)

type Status int
const (
	Good Status = iota
	Bad
	Gone
)

type tributary struct {
	uri      string
	feed     *feed.Feed
	client   *http.Client
	onUpdate []func(models.Feed)
	onStatus []func(Status)
	quit     chan struct{}
}

func newTributary(store persistence.Bucket, uri string, cacheTimeout time.Duration) *tributary {
	p := &tributary{}
	p.uri = uri
	p.feed = feed.New(cacheTimeout, p.chanHandler, p.itemHandler, store)
	p.client = &http.Client{Timeout: time.Minute, Transport: &statusTransport{http.DefaultTransport.(*http.Transport), p}}
	p.onUpdate = []func(models.Feed){}
	p.quit = make(chan struct{})

	go p.poll()
	return p
}

func (t *tributary) OnUpdate(f func(models.Feed)) {
	t.onUpdate = append(t.onUpdate, f)
}

func (t *tributary) OnStatus(f func(Status)) {
	t.onStatus = append(t.onStatus, f)
}

func (t *tributary) Uri() string {
	return t.uri
}

func (t *tributary) poll() {
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

func charsetReader(name string, r io.Reader) (io.Reader, error) {
	return charset.NewReader(name, r)
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

// fetch retrieves the feed for the tributary, if an error occurs or the status
// code is not 200 OK any listeners are notified of the status change.
func (t *tributary) fetch() {
	code, err := t.feed.Fetch(t.uri, t.client, charsetReader)
	if err != nil {
		t.status(Bad)
		log.Println("error fetching", t.uri+":", code, err)
		return
	}

	switch code {
	case http.StatusOK:
		t.status(Good)
	case http.StatusNotModified:
		// ignore
	case http.StatusGone:
		t.status(Gone)
	default:
		t.status(Bad)
	}
}

func (t *tributary) chanHandler(feed *feed.Feed, newchannels []*feed.Channel) {}

func (t *tributary) itemHandler(feed *feed.Feed, ch *feed.Channel, newitems []*feed.Item) {
	items := []models.Item{}
	for _, item := range newitems {
		converted := convertItem(item)

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

func (t *tributary) status(code Status) {
	for _, f := range t.onStatus {
		f(code)
	}
}

func (t *tributary) Kill() {
	close(t.quit)
}
