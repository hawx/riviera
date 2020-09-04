// Package riverjs contains structs that map to the parts of a riverjs feed.
//
// Commentary for the types in this package is copied from http://riverjs.org.
package riverjs

import (
	"strings"
)

type River struct {
	UpdatedFeeds Feeds    `json:"updatedFeeds"`
	Metadata     Metadata `json:"metadata"`
}

type Feeds struct {
	UpdatedFeeds []Feed `json:"updatedFeed"`
}

// Each of these elements are mandatory, if there is no value for one it must be
// included with an empty string as its value.
//
// The elements all come from the top level of a feed, except for FeedURL which
// is the address of the feed itself, and WhenLastUpdate which is the time when
// the new items from the feed were read by the aggregator.
type Feed struct {
	FeedURL         string  `json:"feedUrl"`
	WebsiteURL      string  `json:"websiteUrl"`
	FeedTitle       string  `json:"feedTitle"`
	FeedDescription string  `json:"feedDescription"`
	WhenLastUpdate  RssTime `json:"whenLastUpdate"`
	Items           []Item  `json:"item"`
}

type Item struct {
	// Body is the description from the feed, with html markup stripped, and
	// limited to 280 characters. If the original text was more than the maximum
	// length, three periods are added to the end.
	Body string `json:"body"`

	// Permalink, PubDate, Title and Link are straightforward copies of what
	// appeared in the feed.
	PermaLink string  `json:"permaLink"`
	PubDate   RssTime `json:"pubDate"`
	Title     string  `json:"title"`
	Link      string  `json:"link"`

	// ID is a number assigned to the item by the aggregator. Usuaully it is
	// incremented by one for each item, but that's not guaranteed.
	ID string `json:"id"`

	// Comments points to a page of comments related to the item (it's exactly as
	// in RSS 2.0).
	Comments string `json:"comments,omitempty"`

	// Enclosure is exactly as in RSS 2.0, with three sub-elements, url, type and
	// length.
	Enclosures []Enclosure `json:"enclosure,omitempty"`

	// Thumbnail has three sub-elements, url that points to the full image, and
	// width and height which give the size of the thumbnail.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
}

type Enclosure struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Length int64  `json:"length"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Height *int   `json:"height,omitempty"`
	Width  *int   `json:"width,omitempty"`
}

type Metadata struct {
	// Docs is a link to a web page that documents the format.
	Docs string `json:"docs"`

	// WhenGMT says when the file was built in a universal time.
	WhenGMT RssTime `json:"whenGMT"`

	// WhenLocal says when the file was built in local time.
	WhenLocal RssTime `json:"whenLocal"`

	// Version is 3.
	Version string `json:"version"`

	// Secs is the number of seconds it took to build the file.
	Secs float64 `json:"secs,string"`
}

func (r Item) FilteredBody() string {
	r.Body = strings.TrimSpace(r.Body)

	if strings.HasPrefix(r.Body, "&amp;lt;") ||
		strings.HasPrefix(r.Body, "var gaJsHost") {
		return ""
	}

	return r.Body
}
