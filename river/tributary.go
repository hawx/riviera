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

type Tributary interface {
	OnUpdate(func(models.Feed))
	Uri() string
	Kill()
}

type tributary struct {
	uri      string
	feed     *feed.Feed
	onUpdate []func(models.Feed)
	quit     chan struct{}
}

func newTributary(store persistence.Bucket, uri string, cacheTimeout time.Duration) Tributary {
	p := &tributary{}
	p.uri = uri
	p.feed = feed.New(cacheTimeout, true, p.chanHandler, p.itemHandler, store)
	p.onUpdate = []func(models.Feed){}
	p.quit = make(chan struct{})

	go p.poll()
	return p
}

func (t *tributary) OnUpdate(f func(models.Feed)) {
	t.onUpdate = append(t.onUpdate, f)
}

func (t *tributary) Uri() string {
	return t.uri
}

func (w *tributary) poll() {
	w.fetch()

loop:
	for {
		select {
		case <-w.quit:
			break loop
		case <-time.After(w.feed.DurationTillUpdate()):
			log.Println("fetching", w.uri)
			w.fetch()
		}
	}

	log.Println("stopped fetching", w.uri)
}

func charsetReader(name string, r io.Reader) (io.Reader, error) {
	return charset.NewReader(name, r)
}

func (w *tributary) fetch() {
	if err := w.feed.FetchClient(w.uri, &http.Client{Timeout: time.Minute}, charsetReader); err != nil {
		log.Println("error fetching", w.uri+":", err)
	}
}

func (w *tributary) chanHandler(feed *feed.Feed, newchannels []*feed.Channel) {}

func (w *tributary) itemHandler(feed *feed.Feed, ch *feed.Channel, newitems []*feed.Item) {
	items := []models.Item{}
	for _, item := range newitems {
		converted := convertItem(item)

		if converted != nil {
			items = append(items, *converted)
		}
	}

	log.Println(len(items), "new item(s) in", feed.Url)
	if len(items) == 0 {
		return
	}

	feedUrl := feed.Url
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

	w.notify(models.Feed{
		FeedUrl:         feedUrl,
		WebsiteUrl:      websiteUrl,
		FeedTitle:       ch.Title,
		FeedDescription: ch.Description,
		WhenLastUpdate:  models.RssTime{time.Now()},
		Items:           items,
	})
}

func (w *tributary) notify(feed models.Feed) {
	for _, f := range w.onUpdate {
		f(feed)
	}
}

func (w *tributary) Kill() {
	close(w.quit)
}
