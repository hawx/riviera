package river

import "time"

type Confluence interface {
	Latest() []Feed
	Add(Tributary)
	Remove(string) bool
}

type confluence struct {
	streams []Tributary
	latest  []Feed
}

func newConfluence(streams []Tributary) Confluence {
	c := &confluence{streams, []Feed{}}
	for _, r := range streams {
		c.run(r)
	}
	return c
}

func (c *confluence) run(r Tributary) {
	go func(in <-chan Feed) {
		for v := range in {
			c.latest = append([]Feed{v}, c.latest...)
		}
	}(r.Latest())
}

func (c *confluence) Latest() []Feed {
	yesterday := time.Now().Add(-24 * time.Hour)
	newLatest := []Feed{}

	for _, feed := range c.latest {
		if feed.WhenLastUpdate.After(yesterday) {
			newLatest = append(newLatest, feed)
		}
	}

	c.latest = newLatest
	return c.latest
}

func (c *confluence) Add(stream Tributary) {
	c.streams = append(c.streams, stream)
	c.run(stream)
}

func (c *confluence) Remove(uri string) bool {
	streams := []Tributary{}
	ok := false

	for _, stream := range c.streams {
		if stream.Uri() != uri {
			streams = append(streams, stream)
		} else {
			ok = true
			stream.Kill()
		}
	}

	c.streams = streams
	return ok
}
