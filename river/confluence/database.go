package confluence

import (
	"hawx.me/code/riviera/river/riverjs"

	"time"
)

// A Database contains persisted feed data, specifically each "block" of updates
// for a feed. This allows the river to be recreated from past data, to be
// displayed on startup.
type Database interface {
	Add(feed riverjs.Feed)
	Truncate(cutoff time.Duration)
	Latest(cutoff time.Duration) []riverjs.Feed
}
