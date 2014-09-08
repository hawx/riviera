package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"time"
	"log"
)

type poller struct {
	uri    string
	latest []Feed
	quit   chan struct{}
	feed   *rss.Feed
	cutOff time.Duration
	seen   map[string] struct{}
}

func newPoller(uri string, cutOff time.Duration) River {
	q := make(chan struct{})
	p := &poller{
	  uri: uri,
	  latest: []Feed{},
	  quit: q,
	  feed: nil,
	  cutOff: cutOff,
	  seen: map[string] struct{}{},
	}

	p.feed = rss.New(5, true, p.chanHandler, p.itemHandler)
	go p.poll()
	return p
}

func (w *poller) poll() {
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
}

func (w *poller) fetch() {
	if err := w.feed.Fetch(w.uri, nil); err != nil {
		log.Println("error fetching", w.uri + ":", err)
	}
}

func (w *poller) chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	// ignore
}

func (w *poller) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	items := []Item{}
	for _, item := range newitems {
		converted := convertItem(item)

		if converted == nil { continue }
		if _, seen := w.seen[converted.Id]; seen { continue }

		items = append(items, *converted)
		w.seen[converted.Id] = struct{}{}
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

	w.latest = append(w.latest, Feed{
  	FeedUrl: feedUrl,
    WebsiteUrl: websiteUrl,
	  FeedTitle: ch.Title,
  	FeedDescription: ch.Description,
	  WhenLastUpdate: RssTime{time.Now()},
	  Items: items,
	})
}

func (w *poller) Latest() []Feed {
	filtered := []Feed{}
	for _, feed := range w.latest {
		if feed.WhenLastUpdate.After(time.Now().Add(w.cutOff)) {
			filtered = append(filtered, feed)
		}
	}

	w.latest = filtered
	return w.latest
}

func (w *poller) Close() {
	w.quit <- struct{}{}
}
