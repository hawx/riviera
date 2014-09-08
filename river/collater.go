package river

type Aggregator interface {
	Latest() []Feed
}

type aggregator struct {
	rivers  []River
	latest  []Feed
}

func newAggregator(rivers []River) Aggregator {
	agg := &aggregator{rivers, []Feed{}}
	go agg.aggregate()
	return agg
}

func (c *aggregator) aggregate() {
	for _, r := range c.rivers {
		go func(in <-chan Feed) {
			for v := range in {
				c.latest = append([]Feed{v}, c.latest...)
			}
		}(r.Latest())
	}
}

func (r *aggregator) Latest() []Feed {
	return r.latest
}
