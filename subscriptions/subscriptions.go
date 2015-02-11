// Package subscriptions implements a list of feeds along with operations to
// modify the list.
package subscriptions

import (
	"github.com/hawx/riviera/data"
	"github.com/hawx/riviera/subscriptions/opml"

	"encoding/json"
)

type Status string

const (
	Good Status = "Good"
	Bad  Status = "Bad"
	Gone Status = "Gone"
)

type Subscriptions interface {
	// The list of feeds subscribed to.
	List() []Subscription

	// Add a new feed url to the list.
	Add(string)

	// Refresh the data for a particular Subscription.
	Refresh(Subscription)

	// Remove the Subscription with url provided.
	Remove(string)

	// Call the associated function whenever Add(string) is called, with the
	// Subscription that is created..
	OnAdd(func(Subscription))

	// Call the associated function whenever Remove(string) is called, with the
	// string value provided.
	OnRemove(func(string))
}

// A List provides a read-only view to Subscriptions.
type List interface {
	List() []Subscription
	Refresh(Subscription)
	OnAdd(func(Subscription))
	OnRemove(func(string))
}

type Subscription struct {
	// Uri the subscription was created with, never changed!
	Uri string `json:"uri"`

	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
	Status          Status `json:"status"`
}

type subs struct {
	data.Bucket
	onAdd    []func(Subscription)
	onRemove []func(string)
}

var subscriptionsBucketName = []byte("subscriptions")

// Open loads the Subscriptions from a Database, or initialises an empty set if
// they do not exist.
func Open(db data.Database) (Subscriptions, error) {
	b, err := db.Bucket(subscriptionsBucketName)
	if err != nil {
		return nil, err
	}

	return &subs{b, []func(Subscription){}, []func(string){}}, nil
}

func (s *subs) List() []Subscription {
	subscriptions := []Subscription{}
	s.View(func(tx data.Tx) error {
		for _, e := range tx.All() {
			var s Subscription
			json.Unmarshal(e, &s)
			subscriptions = append(subscriptions, s)
		}
		return nil
	})

	return subscriptions
}

func (s *subs) Add(uri string) {
	sub := Subscription{Uri: uri}
	s.Update(func(tx data.Tx) error {
		value, _ := json.Marshal(sub)

		return tx.Put([]byte(uri), value)
	})

	for _, f := range s.onAdd {
		f(sub)
	}
}

func (s *subs) Refresh(sub Subscription) {
	s.Update(func(tx data.Tx) error {
		value, _ := json.Marshal(sub)

		return tx.Put([]byte(sub.Uri), value)
	})
}

func (s *subs) Remove(uri string) {
	s.Update(func(tx data.Tx) error {
		return tx.Delete([]byte(uri))
	})

	for _, f := range s.onRemove {
		f(uri)
	}
}

func (s *subs) OnAdd(f func(Subscription)) {
	s.onAdd = append(s.onAdd, f)
}

func (s *subs) OnRemove(f func(string)) {
	s.onRemove = append(s.onRemove, f)
}

// FromOpml adds all feeds listed in an opml.Opml document to the Subscriptions.
func FromOpml(s Subscriptions, doc opml.Opml) {
	for _, e := range doc.Body.Outline {
		s.Add(e.XmlUrl)
	}
}

// AsOpml returns a representation of the Subscriptions as an OMPL document.
func AsOpml(s Subscriptions) opml.Opml {
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
