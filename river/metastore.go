package river

import (
	"errors"
	"sort"
)

type Metadata struct {
	Uri             string  `json:"uri"`
	FeedUrl         string  `json:"feedUrl"`
	WebsiteUrl      string  `json:"websiteUrl"`
	FeedTitle       string  `json:"feedTitle"`
	FeedDescription string  `json:"feedDescription"`
	Log             []Event `json:"log"`
}

type feeddata struct {
	Uri             string `json:"uri"`
	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
}

type metaStore struct {
	data  map[string]metaEvent
	evfac func() *events
}

type metaEvent struct {
	meta feeddata
	evs  *events
}

// newMetaStore returns a new, empty, object for storing metadata about feed
// collection. It contains up-to-date information on the feed such as title,
// description, &c. as well as a log of recent fetches. The maximum length of
// this log, per feed, is given by the size argument.
func newMetaStore(size int) *metaStore {
	return &metaStore{
		data:  map[string]metaEvent{},
		evfac: func() *events { return newEvents(size) },
	}
}

func (m *metaStore) Set(key string, meta feeddata) {
	if pair, ok := m.data[key]; ok {
		pair.meta = meta
		m.data[key] = pair
		return
	}

	m.data[key] = metaEvent{
		meta: meta,
		evs:  m.evfac(),
	}
}

func (m *metaStore) Delete(key string) {
	delete(m.data, key)
}

func (m *metaStore) Log(key string, event Event) error {
	if pair, ok := m.data[key]; ok {
		pair.evs.Prepend(event)
		return nil
	}

	return errors.New("key does not exist")
}

func (m *metaStore) List() []Metadata {
	keys := []string{}
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	list := make([]Metadata, len(keys))
	for i, k := range keys {
		thismeta := m.data[k].meta

		list[i] = Metadata{
			Uri:             thismeta.Uri,
			FeedUrl:         thismeta.FeedUrl,
			WebsiteUrl:      thismeta.WebsiteUrl,
			FeedTitle:       thismeta.FeedTitle,
			FeedDescription: thismeta.FeedDescription,
			Log:             m.data[k].evs.List(),
		}
	}
	return list
}
