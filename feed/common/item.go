package common

import (
	"crypto/md5"
	"io"
	"time"
)

type Guid struct {
	Guid        string
	IsPermaLink bool
}

type Item struct {
	// RSS and Shared fields
	Author      Author
	Categories  []Category
	Comments    string
	Description string
	Enclosures  []Enclosure
	Extensions  map[string]map[string][]Extension
	Guid        *Guid
	Links       []Link
	PubDate     string
	Source      *Source
	Title       string

	// Atom specific fields
	Content      *Content
	Contributors []string
	Generator    *Generator
	Id           string
}

func (i *Item) ParsedPubDate() (time.Time, error) {
	return parseTime(i.PubDate)
}

func (i *Item) Key() string {
	switch {
	case i.Guid != nil && len(i.Guid.Guid) != 0:
		return i.Guid.Guid
	case len(i.Id) != 0:
		return i.Id
	case len(i.Title) > 0 && len(i.PubDate) > 0:
		return i.Title + i.PubDate
	default:
		h := md5.New()
		io.WriteString(h, i.Description)
		return string(h.Sum(nil))
	}
}
