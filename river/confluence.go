package river

type Confluence interface {
	Latest() []Feed
}

type confluence struct {
	streams []Tributary
	latest  []Feed
}

func newConfluence(streams []Tributary) Confluence {
	con := &confluence{streams, []Feed{}}
	go con.run()
	return con
}

func (c *confluence) run() {
	for _, r := range c.streams {
		go func(in <-chan Feed) {
			for v := range in {
				c.latest = append([]Feed{v}, c.latest...)
			}
		}(r.Latest())
	}
}

func (r *confluence) Latest() []Feed {
	return r.latest
}
