package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"time"
	"log"
)

type River interface {
	Latest() []Feed
	Close()
}

type collater struct {
	rivers  map[string] River
}

func New(uris []string) River {
	rivers := map[string]River{}

	for _, uri := range uris {
		rivers[uri] = newPoller(uri)
	}

	return &collater{rivers}
}

func (r *collater) Latest() []Feed {
	feeds := []Feed{}

	for _, river := range r.rivers {
		feeds = append(feeds, river.Latest()...)
	}

	return feeds
}

func (r *collater) Close() {
	for _, river := range r.rivers {
		river.Close()
	}
}

type poller struct {
	uri    string
	latest []Feed
	quit   chan struct{}
	feed   *rss.Feed
}

func newPoller(uri string) River {
	q := make(chan struct{})
	p := &poller{uri, []Feed{}, q, nil}
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
		case <-time.After(time.Duration(w.feed.SecondsTillUpdate() * 1e9)):
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
	f := *convertChannel(ch, w.uri, time.Duration(24 * time.Hour))
	log.Println(len(f.Items), "new item(s) in", w.uri)
	w.latest = []Feed{f}
}

func (w *poller) Latest() []Feed {
	return w.latest
}

func (w *poller) Close() {
	w.quit <- struct{}{}
}
