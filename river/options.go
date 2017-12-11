package river

import (
	"time"

	"hawx.me/code/riviera/river/mapping"
)

type Options struct {
	// Mapping is the function used to convert a feed item to an item in the
	// river.
	Mapping mapping.Mapping

	// CutOff is the duration after which items are not shown in the river. This
	// is given as a negative time and is calculated from the time the feed was
	// fetched not the time the item was published.
	CutOff time.Duration

	// Refresh is the minimum refresh period. If an rss feed does not specify
	// when to be fetched this duration will be used.
	Refresh time.Duration

	// LogLength defines the number of events to keep in the crawl log, per feed.
	LogLength int
}

var DefaultOptions = Options{
	Mapping:   mapping.DefaultMapping,
	CutOff:    -24 * time.Hour,
	Refresh:   15 * time.Minute,
	LogLength: 0,
}
