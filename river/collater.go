package river

import "sort"

type collater struct {
	rivers  map[string] River
}

func (r *collater) Latest() []Feed {
	feeds := []Feed{}

	for _, river := range r.rivers {
		feeds = append(feeds, river.Latest()...)
	}

	sort.Sort(ByWhenLastUpdate(feeds))
	return feeds
}

func (r *collater) Close() {
	for _, river := range r.rivers {
		river.Close()
	}
}

type ByWhenLastUpdate []Feed

func (a ByWhenLastUpdate) Len() int {
	return len(a)
}

func (a ByWhenLastUpdate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByWhenLastUpdate) Less(i, j int) bool {
	return a[i].WhenLastUpdate.Time.After(a[j].WhenLastUpdate.Time)
}
