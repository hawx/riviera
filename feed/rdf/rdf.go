package rdf

// http://web.resource.org/rss/1.0/spec

import (
	"encoding/xml"
	"io"

	"hawx.me/code/riviera/feed/data"
)

var days = map[string]int{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    7,
}

type Parser struct{}

func (Parser) CanRead(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) bool {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset

	var token xml.Token
	var err error
	for {
		if token, err = decoder.Token(); err != nil || token == nil {
			return false
		}

		if t, ok := token.(xml.StartElement); ok {
			return t.Name.Space == "http://www.w3.org/1999/02/22-rdf-syntax-ns#" && t.Name.Local == "RDF"
		}
	}

	return false
}

func (Parser) Read(r io.Reader, charset func(string, io.Reader) (io.Reader, error)) (foundChannels []*data.Channel, err error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset

	var feed rdfFeed
	if err = decoder.Decode(&feed); err != nil {
		return
	}

	ch := &data.Channel{
		Title:       feed.Channel.Title,
		Description: feed.Channel.Description,
		Links: []data.Link{
			data.Link{Href: feed.Channel.Link},
		},
	}

	if feed.Channel.Image != nil {
		ch.Image = data.Image{
			Title: feed.Image.Title,
			Url:   feed.Image.URL,
			Link:  feed.Image.Link,
		}
	}

	return
}

type rdfFeed struct {
	XMLName xml.Name `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`

	Channel rdfChannel    `xml:"channel"`
	Image   *rdfFeedImage `xml:"image"`
	Items   []rdfItem     `xml:"item"`
}

type rdfChannel struct {
	About       string    `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Image       *rdfImage `xml:"image"`
	Items       rdfItems  `xml:"items"`
}

type rdfImage struct {
	Resource string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# resource,attr"`
}

type rdfItems struct {
	Seq rdfSeq `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# Seq"`
}

type rdfSeq struct {
	Li rdfLi `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# li"`
}

type rdfLi struct {
	Resource string `xml:"resource,attr"`
}

type rdfFeedImage struct {
	About string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
	URL   string `xml:"url"`
}

type rdfItem struct {
	About       string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}
