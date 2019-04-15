package common

import (
	"crypto/md5"
	"io"
	"time"
)

type GUID struct {
	GUID        string
	IsPermaLink bool
}

type Item struct {
	// RSS and Shared fields
	Author      Author
	Categories  []Category
	Comments    string
	Description string
	Enclosures  []Enclosure
	Thumbnail   *Image
	Extensions  map[string]map[string][]Extension
	GUID        *GUID
	Links       []Link
	PubDate     string
	Source      *Source
	Title       string

	// Atom specific fields
	Content      *Content
	Contributors []string
	Generator    *Generator
	ID           string
}

func (i *Item) ParsedPubDate() (time.Time, error) {
	return parseTime(i.PubDate)
}

func (i *Item) Key() string {
	switch {
	case i.GUID != nil && len(i.GUID.GUID) != 0:
		return i.GUID.GUID
	case len(i.ID) != 0:
		return i.ID
	case len(i.Title) > 0 && len(i.PubDate) > 0:
		return i.Title + i.PubDate
	default:
		h := md5.New()
		if _, err := io.WriteString(h, i.Description); err != nil {
			panic(err)
		}
		return string(h.Sum(nil))
	}
}
