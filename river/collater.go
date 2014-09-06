package river

type collater struct {
	rivers  map[string] River
}

func (r *collater) Latest() []Feed {
	feeds := []Feed{}

	for _, river := range r.rivers {
		feeds = append(feeds, river.Latest()...)
	}

	return feeds
}

func (r *collater) Close() {
	for _, river := range r.rivers {
		river.Close()
	}
}
