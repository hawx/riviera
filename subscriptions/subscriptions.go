package subscriptions

import (
	"github.com/hawx/riviera/subscriptions/opml"
)

type Subscriptions interface {
	List() []string
	Add(string)
	Remove(string)
	Events() <-chan Event
}

type List interface {
	List() []string
	Events() <-chan Event
}

type Event struct {
	Type EventType
	Uri  string
}

type EventType int

const (
	Add = iota
	Remove
)

type subscriptions struct {
	path   string
	subs   *opml.Opml
	events chan Event
}

func Load(path string) (Subscriptions, error) {
	subs, err := opml.Load(path)
	if err != nil {
		return nil, err
	}

	return &subscriptions{path, subs, make(chan Event)}, nil
}

func (s *subscriptions) List() []string {
	urls := []string{}
	for _, outline := range s.subs.Body.Outline {
		urls = append(urls, outline.XmlUrl)
	}

	return urls
}

func (s *subscriptions) Add(url string) {
	s.subs.Body.Outline = append(s.subs.Body.Outline, opml.Outline{XmlUrl: url})
	s.subs.Save(s.path)
	s.events <- Event{Add, url}
}

func (s *subscriptions) Remove(url string) {
	body := []opml.Outline{}
	for _, outline := range s.subs.Body.Outline {
		if outline.XmlUrl != url {
			body = append(body, outline)
		}
	}

	s.subs.Body.Outline = body
	s.subs.Save(s.path)
	s.events <- Event{Remove, url}
}

func (s *subscriptions) Events() <-chan Event {
	return s.events
}
