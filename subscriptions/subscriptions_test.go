package subscriptions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/subscriptions/opml"
)

func TestSubscriptions(t *testing.T) {
	subs := New()

	assert.Equal(t, []Subscription(nil), subs.List())

	// Add feed
	subs.Add("http://example.com/feed")
	subs.Add("http://example.org/xml")
	subs.Add("http://example.com/feed2")

	assert.Equal(t, []Subscription{
		{URI: "http://example.com/feed"},
		{URI: "http://example.com/feed2"},
		{URI: "http://example.org/xml"},
	}, subs.List())

	// Refresh feed
	subs.Refresh(Subscription{
		URI:       "http://example.com/feed",
		FeedTitle: "what",
	})
	assert.Equal(t, []Subscription{
		{URI: "http://example.com/feed", FeedTitle: "what"},
		{URI: "http://example.com/feed2"},
		{URI: "http://example.org/xml"},
	}, subs.List())

	// Remove feed
	subs.Remove("http://example.com/feed")
	assert.Equal(t, []Subscription{
		{URI: "http://example.com/feed2"},
		{URI: "http://example.org/xml"},
	}, subs.List())
}

func TestDiffWhenChanged(t *testing.T) {
	a := New()
	a.Add("http://example.com/feed2")
	a.Add("http://example.com/feed")

	b := New()
	b.Add("http://example.com/feed")
	b.Add("http://example.org/xml")

	added, removed := Diff(a, b)
	assert.Equal(t, []string{"http://example.org/xml"}, added)
	assert.Equal(t, []string{"http://example.com/feed2"}, removed)
}

func TestFromOpml(t *testing.T) {
	doc := opml.Opml{
		Version: "1.1",
		Body: opml.Body{Outline: []opml.Outline{
			{ // ignored as type not "rss"
				Type:   "whu",
				Text:   "hey",
				XMLURL: "what",
			},
			{
				Type:   "rss",
				Text:   "hey2",
				XMLURL: "what2",
			},
			{
				Type:        "rss",
				Text:        "cool",
				XMLURL:      "yes",
				Description: "this desc",
				HTMLURL:     "htmls",
				Language:    "en",
				Title:       "titl",
			},
		}},
	}

	subs := FromOpml(doc)

	assert.Equal(t, []Subscription{
		{URI: "what2", FeedTitle: "hey2", FeedURL: "what2"},
		{URI: "yes", FeedTitle: "cool", FeedURL: "yes", WebsiteURL: "htmls", FeedDescription: "this desc"},
	}, subs.List())
}
