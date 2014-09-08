// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	"encoding/json"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type River interface {
	Latest() []Feed
	Close()
}

func New(uris []string, cutOff time.Duration) River {
	rivers := map[string]River{}

	for _, uri := range uris {
		rivers[uri] = newPoller(uri, cutOff)
	}

	return &collater{rivers}
}

func Build(river River) string {
	return fromFeeds(river.Latest())
}

func fromFeeds(feeds []Feed) string {
	updatedFeeds := Feeds{feeds}
	now := time.Now()

	metadata := Metadata{
		Docs:      DOCS,
	  WhenGMT:   RssTime{now.UTC()},
		WhenLocal: RssTime{now},
		Version:   "3",
		Secs:      0,
	}

	wrapper := Wrapper{
		Metadata:     metadata,
		UpdatedFeeds: updatedFeeds,
	}

	b, _ := json.MarshalIndent(wrapper, "", "  ")

	return string(b)
}
