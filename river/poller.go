package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"time"
	"log"
)

type poller struct {
	uri    string
	latest []Feed
	cutOff time.Duration
	quit   chan struct{}
	feed   *rss.Feed
}

func newPoller(uri string, cutOff time.Duration) River {
	q := make(chan struct{})
	p := &poller{uri, []Feed{}, cutOff, q, nil}
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
	f := *convertChannel(ch, w.uri, w.cutOff)
	log.Println(len(f.Items), "new item(s) in", w.uri)
	w.latest = []Feed{f}
}

func (w *poller) Latest() []Feed {
	return w.latest
}

func (w *poller) Close() {
	w.quit <- struct{}{}
}
