package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"time"
	"log"
)

type River interface {
	Latest() <-chan Feed
}

type poller struct {
	uri    string
	feed   *rss.Feed
	in     chan Feed
}

func newPoller(uri string) River {
	p := &poller{}
	p.uri = uri
	p.in = make(chan Feed)
	p.feed = rss.New(5, true, p.chanHandler, p.itemHandler)

	go p.poll()
	return p
}

func (w *poller) poll() {
	w.fetch()

	for {
		select {
		case <-time.After(time.Duration(w.feed.SecondsTillUpdate()) * time.Second):
			w.fetch()
		}
	}
}

func (w *poller) fetch() {
	if err := w.feed.Fetch(w.uri, nil); err != nil {
		log.Println("error fetching", w.uri + ":", err)
	}
}

func (w *poller) chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {}

func (w *poller) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	items := []Item{}
	for _, item := range newitems {
		converted := convertItem(item)

		if converted != nil {
			items = append(items, *converted)
		}
	}

	log.Println(len(items), "new item(s) in", feed.Url)
	if len(items) == 0 { return }

	feedUrl := feed.Url
	websiteUrl := ""
	for _, link := range ch.Links {
		if feedUrl != "" && websiteUrl != "" { break }

		if link.Rel == "self" {
			feedUrl = link.Href
		} else {
			websiteUrl = link.Href
		}
	}

	w.in <- Feed{
  	FeedUrl: feedUrl,
    WebsiteUrl: websiteUrl,
	  FeedTitle: ch.Title,
  	FeedDescription: ch.Description,
	  WhenLastUpdate: RssTime{time.Now()},
	  Items: items,
	}
}

func (w *poller) Latest() <-chan Feed {
	return w.in
}
