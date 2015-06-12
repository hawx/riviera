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
		{Uri: "http://example.com/feed"},
		{Uri: "http://example.com/feed2"},
		{Uri: "http://example.org/xml"},
	}, subs.List())

	// Refresh feed
	subs.Refresh(Subscription{
		Uri:       "http://example.com/feed",
		FeedTitle: "what",
	})
	assert.Equal(t, []Subscription{
		{Uri: "http://example.com/feed", FeedTitle: "what"},
		{Uri: "http://example.com/feed2"},
		{Uri: "http://example.org/xml"},
	}, subs.List())

	// Remove feed
	subs.Remove("http://example.com/feed")
	assert.Equal(t, []Subscription{
		{Uri: "http://example.com/feed2"},
		{Uri: "http://example.org/xml"},
	}, subs.List())
}

func TestDiffWhenChanged(t *testing.T) {
	a := New()
	a.Add("http://example.com/feed2")
	a.Add("http://example.com/feed")

	b := New()
	b.Add("http://example.com/feed")
	b.Add("http://example.org/xml")

	assert.Equal(t, []Change{
		Change{Removed, "http://example.com/feed2"},
		Change{Added, "http://example.org/xml"},
	}, Diff(a, b))
}

func TestFromOpml(t *testing.T) {
	doc := opml.Opml{
		Version: "1.1",
		Body: opml.Body{Outline: []opml.Outline{
			{ // ignored as type not "rss"
				Type:   "whu",
				Text:   "hey",
				XmlUrl: "what",
			},
			{
				Type:   "rss",
				Text:   "hey2",
				XmlUrl: "what2",
			},
			{
				Type:        "rss",
				Text:        "cool",
				XmlUrl:      "yes",
				Description: "this desc",
				HtmlUrl:     "htmls",
				Language:    "en",
				Title:       "titl",
			},
		}},
	}

	subs := FromOpml(doc)

	assert.Equal(t, []Subscription{
		{Uri: "what2", FeedTitle: "hey2", FeedUrl: "what2"},
		{Uri: "yes", FeedTitle: "cool", FeedUrl: "yes", WebsiteUrl: "htmls", FeedDescription: "this desc"},
	}, subs.List())
}
