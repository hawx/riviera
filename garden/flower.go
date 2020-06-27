package garden

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
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

	items gardenjs.Feed
}

func NewFlower(store feed.Database, cacheTimeout time.Duration, uri string, size int) (*Flower, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	f := &Flower{
		uri:    parsedURI,
		client: http.DefaultClient,
		quit:   make(chan struct{}),
		size:   size,
	}
	f.feed = feed.New(cacheTimeout, f.itemHandler, store)

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

	for _, item := range newitems {
		converted := mapping.DefaultMapping(item)

		if converted != nil {
			converted.Link = maybeResolvedLink(f.uri, converted.Link)
			converted.PermaLink = maybeResolvedLink(f.uri, converted.PermaLink)

			f.items.Items = append([]gardenjs.Item{{
				PermaLink: converted.PermaLink,
				PubDate:   converted.PubDate.Add(0),
				Title:     converted.Title,
				Link:      converted.Link,
			}}, f.items.Items...)
		}
	}

	if len(f.items.Items) > f.size {
		f.items.Items = f.items.Items[:f.size]
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

	f.items.URL = feedURL
	f.items.WebsiteURL = websiteURL
	f.items.Title = ch.Title
	f.items.UpdatedAt = time.Now()
}

func maybeResolvedLink(root *url.URL, other string) string {
	parsed, err := root.Parse(other)
	if err == nil {
		return parsed.String()
	}

	return other
}
