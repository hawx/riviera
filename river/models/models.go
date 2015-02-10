package models

import "time"

// Commentary copied from riverjs.org.

type Wrapper struct {
	UpdatedFeeds Feeds    `json:"updatedFeeds"`
	Metadata     Metadata `json:"metadata"`
}

type Feeds struct {
	UpdatedFeeds []Feed `json:"updatedFeed"`
}

// Each of these elements are mandatory, if there is no value for one it must be
// included with an empty string as its value.
//
// The elements all come from the top level of a feed, except for FeedUrl which
// is the address of the feed itself, and WhenLastUpdate which is the time when
// the new items from the feed were read by the aggregator.
type Feed struct {
	FeedUrl         string  `json:"feedUrl"`
	WebsiteUrl      string  `json:"websiteUrl"`
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

	// Id is a number assigned to the item by the aggregator. Usuaully it is
	// incremented by one for each item, but that's not guaranteed.
	Id string `json:"id"`

	// Comments points to a page of comments related to the item (it's exactly as
	// in RSS 2.0).
	Comments string `json:"comments,omitempty"`

	// Enclosure is exactly as in RSS 2.0, with three sub-elements, url, type and
	// length.
	Enclosures []Enclosure `json:"enclosure,omitempty"`
}

type Enclosure struct {
	Url    string `json:"url"`
	Type   string `json:"type"`
	Length int64  `json:"length"`
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

type RssTime struct {
	time.Time
}

func (t RssTime) MarshalText() ([]byte, error) {
	return []byte(t.Format(time.RFC1123Z)), nil
}

func (t RssTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(time.RFC1123Z) + `"`), nil
}

func (t *RssTime) UnmarshalText(data []byte) error {
	g, err := time.Parse(time.RFC1123Z, string(data))
	if err != nil {
		return err
	}
	*t = RssTime{g}
	return nil
}

func (t *RssTime) UnmarshalJSON(data []byte) error {
	g, err := time.Parse(`"`+time.RFC1123Z+`"`, string(data))
	if err != nil {
		return err
	}
	*t = RssTime{g}
	return nil
}
