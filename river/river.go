// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	"github.com/hawx/riviera/river/database"
	"encoding/json"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

func New(uris []string, store database.Master, cutOff time.Duration) Confluence {
	streams := make([]Tributary, len(uris))

	for i, uri := range uris {
		streams[i] = newTributary(store.Bucket(uri), uri)
	}

	return newConfluence(streams)
}

func Build(river Confluence) string {
	updatedFeeds := Feeds{river.Latest()}
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

func Add(river Confluence, store database.Master, uri string) {
	river.Add(newTributary(store.Bucket(uri), uri))
}

func Remove(river Confluence, uri string) bool {
	return river.Remove(uri)
}
