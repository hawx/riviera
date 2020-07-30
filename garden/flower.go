package garden

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
	"hawx.me/code/riviera/data"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/feed/common"
	"hawx.me/code/riviera/garden/gardenjs"
	"hawx.me/code/riviera/river/mapping"
)

type Flower struct {
	uri    *url.URL
	feed   *feed.Feed
	client *http.Client
	size   int
	quit   chan struct{}
	db     Database

	items gardenjs.Feed
}

type dbWrapper struct {
	db  Database
	uri string
}

func (d *dbWrapper) Contains(key string) bool {
	return d.db.Contains(d.uri, key)
}

func NewFlower(db Database, cacheTimeout time.Duration, uri string, size int) (*Flower, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	f := &Flower{
		uri:    parsedURI,
		client: http.DefaultClient,
		quit:   make(chan struct{}),
		size:   size,
		db:     db,
	}

	f.feed = feed.New(cacheTimeout, f.itemHandler, &dbWrapper{db: db, uri: uri})

	return f, nil
}

func (f *Flower) Latest() gardenjs.Feed {
	return f.items
}

func (f *Flower) Start() {
	go func() {
		log.Println("started fetching", f.uri)
		code, err := f.feed.Fetch(f.uri.String(), f.client, charset.NewReaderLabel)
		if err != nil {
			log.Printf("error fetching %s: %d %s\n", f.uri, code, err)
		}

	loop:
		for {
			select {
			case <-time.After(f.feed.DurationTillUpdate()):
				code, err := f.feed.Fetch(f.uri.String(), f.client, charset.NewReaderLabel)
				if err != nil {
					log.Printf("error fetching %s: %d %s\n", f.uri, code, err)
				}

			case <-f.quit:
				break loop
			}
		}

		close(f.quit)
	}()
}

func (f *Flower) Stop() {
	f.quit <- struct{}{}
	<-f.quit
}

func (f *Flower) itemHandler(feed *feed.Feed, ch *common.Channel, newitems []*common.Item) {
	if len(newitems) == 0 {
		return
	}

	items := make([]data.FeedItem, len(newitems))

	for i, item := range newitems {
		converted := mapping.DefaultMapping(item)

		if converted != nil {
			converted.Link = maybeResolvedLink(f.uri, converted.Link)
			converted.PermaLink = maybeResolvedLink(f.uri, converted.PermaLink)

			items[i] = data.FeedItem{
				Key:       converted.PermaLink,
				PermaLink: converted.PermaLink,
				PubDate:   converted.PubDate.Add(0),
				Title:     converted.Title,
				Link:      converted.Link,
			}
		}
	}

	feedURL := f.uri.String()
	websiteURL := ""
	for _, link := range ch.Links {
		if feedURL != "" && websiteURL != "" {
			break
		}

		if link.Rel == "self" {
			feedURL = maybeResolvedLink(f.uri, link.Href)
		} else {
			websiteURL = maybeResolvedLink(f.uri, link.Href)
		}
	}

	log.Println("updating feed")
	if err := f.db.UpdateFeed(data.Feed{
		FeedURL:     feedURL,
		WebsiteURL:  websiteURL,
		Title:       ch.Title,
		Description: ch.Description,
		UpdatedAt:   time.Now(),
		Items:       items,
	}); err != nil {
		log.Println(err)
	}
}

func maybeResolvedLink(root *url.URL, other string) string {
	parsed, err := root.Parse(other)
	if err == nil {
		return parsed.String()
	}

	return other
}
