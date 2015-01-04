package subscriptions

import (
	"github.com/hawx/riviera/data"
	"github.com/hawx/riviera/subscriptions/opml"

	"encoding/json"
)

type Subscriptions interface {
	List() []Subscription
	Import(*opml.Opml)
	Add(string)
	Remove(string)
	OnAdd(Callback)
	OnRemove(Callback)
}

type List interface {
	List() []Subscription
	OnAdd(Callback)
	OnRemove(Callback)
}

type Callback func(string)

type Subscription struct {
	// Uri the subscription was created with, never changed
	Uri string `json:"uri"`

	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
}

type subs struct {
	data.Bucket
	onAdd    []Callback
	onRemove []Callback
}

var subscriptionsBucketName = []byte("subscriptions")

func Open(db data.Database) (Subscriptions, error) {
	b, err := db.Bucket(subscriptionsBucketName)
	if err != nil {
		return nil, err
	}

	return &subs{b, []Callback{}, []Callback{}}, nil
}

func (s *subs) Import(outline *opml.Opml) {
	for _, e := range outline.Body.Outline {
		s.Add(e.XmlUrl)
	}
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
	s.Update(func(tx data.Tx) error {
		value, _ := json.Marshal(Subscription{Uri: uri})

		return tx.Put([]byte(uri), []byte(value))
	})

	for _, f := range s.onAdd {
		f(uri)
	}
}

func (s *subs) Remove(uri string) {
	s.Update(func(tx data.Tx) error {
		return tx.Delete([]byte(uri))
	})

	for _, f := range s.onRemove {
		f(uri)
	}
}

func (s *subs) OnAdd(f Callback) {
	s.onAdd = append(s.onAdd, f)
}

func (s *subs) OnRemove(f Callback) {
	s.onRemove = append(s.onRemove, f)
}

func AsOpml(s Subscriptions) opml.Opml {
	l := opml.Opml{
		Version: "1.1",
		Head:    opml.Head{Title: "Subscriptions"},
		Body:    opml.Body{Outline: []opml.Outline{}},
	}

	for _, e := range s.List() {
		l.Body.Outline = append(l.Body.Outline, opml.Outline{XmlUrl: e.Uri})
	}

	return l
}
