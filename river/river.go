// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	"github.com/hawx/riviera/river/models"
	"github.com/hawx/riviera/river/database"
	"encoding/json"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type River interface {
	Build() string
	Add(uri string)
	Remove(uri string) bool
}

type river struct {
	confluence Confluence
	store      database.Master
}

func New(store database.Master, cutOff time.Duration, uris []string) River {
	streams := make([]Tributary, len(uris))

	for i, uri := range uris {
		streams[i] = newTributary(store.Bucket(uri), uri)
	}

	return &river{newConfluence(store.River(), streams), store}
}

func (r *river) Build() string {
	updatedFeeds := models.Feeds{r.confluence.Latest()}
	now := time.Now()

	metadata := models.Metadata{
		Docs:      DOCS,
	  WhenGMT:   models.RssTime{now.UTC()},
		WhenLocal: models.RssTime{now},
		Version:   "3",
		Secs:      0,
	}

	wrapper := models.Wrapper{
		Metadata:     metadata,
		UpdatedFeeds: updatedFeeds,
	}

	b, _ := json.MarshalIndent(wrapper, "", "  ")
	return string(b)
}

func (r *river) Add(uri string) {
	r.confluence.Add(newTributary(r.store.Bucket(uri), uri))
}

func (r *river) Remove(uri string) bool {
	return r.confluence.Remove(uri)
}
