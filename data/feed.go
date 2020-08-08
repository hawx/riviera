package data

import "time"

type Feed struct {
	FeedURL     string
	WebsiteURL  string
	Title       string
	Description string
	UpdatedAt   time.Time
	Items       []FeedItem
}

type FeedItem struct {
	Key        string
	PermaLink  string
	PubDate    time.Time
	Title      string
	Link       string
	Body       string
	ID         string
	Comments   string
	Enclosures []FeedItemEnclosure
	Thumbnails []FeedItemThumbnail
}

type FeedItemEnclosure struct {
	URL    string
	Type   string
	Length int
}

type FeedItemThumbnail struct {
	URL    string
	Height int
	Width  int
}
