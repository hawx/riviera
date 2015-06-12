// Package subscriptions implements a list of feeds along with operations to
// modify the list.
package subscriptions

import (
	"sort"
	"sync"

	"hawx.me/code/riviera/subscriptions/opml"
)

// A List provides a read-only view to Subscriptions.
type List interface {
	List() []Subscription
	Refresh(Subscription)
}

// Subscription represents the metadata for a single feed.
type Subscription struct {
	// Uri the subscription was created with, never changed!
	Uri string `json:"uri"`

	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
}

// Subscriptions is a list of subscriptions that is safe to access across
// goroutines.
type Subscriptions struct {
	m  map[string]Subscription
	mu sync.RWMutex
}

// New returns an empty subscription list.
func New() *Subscriptions {
	return &Subscriptions{m: map[string]Subscription{}}
}

// List of feeds subscribed to.
func (s *Subscriptions) List() []Subscription {
	var l []Subscription

	s.mu.RLock()
	var keys []string
	for k := range s.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		l = append(l, s.m[k])
	}
	s.mu.RUnlock()

	return l
}

// Add a new feed url to the list.
func (s *Subscriptions) Add(uri string) {
	s.mu.Lock()
	s.m[uri] = Subscription{Uri: uri}
	s.mu.Unlock()
}

// Refresh the data for a particular Subscription.
func (s *Subscriptions) Refresh(sub Subscription) {
	s.mu.Lock()
	s.m[sub.Uri] = sub
	s.mu.Unlock()
}

// Remove the Subscription with url provided.
func (s *Subscriptions) Remove(uri string) {
	s.mu.Lock()
	delete(s.m, uri)
	s.mu.Unlock()
}

// FromOpml adds all feeds listed in an opml.Opml document to the Subscriptions.
func FromOpml(doc opml.Opml) *Subscriptions {
	s := New()
	for _, e := range doc.Body.Outline {
		if e.Type != "rss" {
			continue
		}

		s.Refresh(Subscription{
			FeedTitle:       e.Text,
			FeedUrl:         e.XmlUrl,
			Uri:             e.XmlUrl,
			WebsiteUrl:      e.HtmlUrl,
			FeedDescription: e.Description,
		})
	}
	return s
}

// AsOpml returns a representation of the Subscriptions as an OMPL document.
func AsOpml(s List) opml.Opml {
	l := opml.Opml{
		Version: "1.1",
		Head:    opml.Head{Title: "Subscriptions"},
		Body:    opml.Body{Outline: []opml.Outline{}},
	}

	for _, e := range s.List() {
		l.Body.Outline = append(l.Body.Outline, opml.Outline{
			Type:        "rss",
			Text:        e.FeedTitle,
			XmlUrl:      e.Uri,
			Description: e.FeedDescription,
			HtmlUrl:     e.WebsiteUrl,
			Title:       e.FeedTitle,
		})
	}

	return l
}

type ChangeType int

const (
	Removed ChangeType = iota
	Added
)

type Change struct {
	Type ChangeType
	Uri  string
}

// Diff finds the difference between two subscription lists.
func Diff(a, b *Subscriptions) []Change {
	var changes []Change

	a.mu.RLock()
	b.mu.RLock()

	for _, s := range a.m {
		if _, ok := b.m[s.Uri]; !ok {
			changes = append(changes, Change{Removed, s.Uri})
		}
	}

	for _, s := range b.m {
		if _, ok := a.m[s.Uri]; !ok {
			changes = append(changes, Change{Added, s.Uri})
		}
	}

	a.mu.RUnlock()
	b.mu.RUnlock()

	return changes
}
