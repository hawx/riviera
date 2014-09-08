// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	"encoding/json"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

func New(uris []string, cutOff time.Duration) Aggregator {
	rivers := make([]River, len(uris))

	for i, uri := range uris {
		rivers[i] = newPoller(uri)
	}

	return newAggregator(rivers)
}

func Build(river Aggregator) string {
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
