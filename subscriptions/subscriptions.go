package subscriptions

import (
	"github.com/hawx/riviera/subscriptions/opml"
)

type Subscriptions interface {
	List() []string
	Add(string)
	Remove(string)
}

type subscriptions struct {
	path string
	subs *opml.Opml
}

func Load(path string) (Subscriptions, error) {
	subs, err := opml.Load(path)
	if err != nil {
		return nil, err
	}

	return &subscriptions{path, subs}, nil
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
}
