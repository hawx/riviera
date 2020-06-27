// Package gardenjs defines types that build gardenjs format feeds.
package gardenjs

import "time"

type Garden struct {
	Feeds    []Feed   `json:"feeds"`
	Metadata Metadata `json:"metadata"`
}

type Feed struct {
	URL        string    `json:"url"`
	WebsiteURL string    `json:"websiteUrl"`
	Title      string    `json:"title"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Items      []Item    `json:"items"`
}

type Item struct {
	PermaLink string    `json:"permaLink"`
	PubDate   time.Time `json:"pubDate"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
}

type Metadata struct {
	BuiltAt time.Time `json:"builtAt"`
}
