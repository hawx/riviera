package river

import (
	rss "github.com/hawx/go-pkg-rss"
	"github.com/hawx/riviera/river/database"
	"github.com/hawx/riviera/river/models"
	"log"
	"net/http"
	"time"
)

type Tributary interface {
	Latest() <-chan models.Feed
	Uri() string
	Kill()
}

type tributary struct {
	uri  string
	feed *rss.Feed
	in   chan models.Feed
	quit chan struct{}
}

func newTributary(store database.Bucket, uri string) Tributary {
	p := &tributary{}
	p.uri = uri
	p.feed = rss.New(5, true, p.chanHandler, p.itemHandler, store)
	p.in = make(chan models.Feed)
	p.quit = make(chan struct{})

	go p.poll()
	return p
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
		case <-time.After(time.Duration(w.feed.SecondsTillUpdate()) * time.Second):
			w.fetch()
		}
	}

	log.Println("stopped fetching", w.uri)
}

func (w *tributary) fetch() {
	if err := w.feed.FetchClient(w.uri, &http.Client{Timeout: time.Minute}, nil); err != nil {
		log.Println("error fetching", w.uri+":", err)
	}
}

func (w *tributary) chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {}

func (w *tributary) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
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

	w.in <- models.Feed{
		FeedUrl:         feedUrl,
		WebsiteUrl:      websiteUrl,
		FeedTitle:       ch.Title,
		FeedDescription: ch.Description,
		WhenLastUpdate:  models.RssTime{time.Now()},
		Items:           items,
	}
}

func (w *tributary) Latest() <-chan models.Feed {
	return w.in
}

func (w *tributary) Kill() {
	close(w.in)
	close(w.quit)
}
