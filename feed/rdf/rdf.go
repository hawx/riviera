// Package rdf provides a parser for RDF Site Summary (RSS) 1.0
//
// See http://web.resource.org/rss/1.0/spec for the specification.
//
// It also supports three modules:
//   - Dublin Core: http://web.resource.org/rss/1.0/modules/dc/
//   - Syndication: http://web.resource.org/rss/1.0/modules/syndication/
//   - Content: http://web.resource.org/rss/1.0/modules/content/
package rdf

import (
	"encoding/xml"
	"io"
	"time"

	"hawx.me/code/riviera/feed/common"
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

func (Parser) Read(r io.Reader, charset func(string, io.Reader) (io.Reader, error)) (foundChannels []*common.Channel, err error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset

	var feed rdfFeed
	if err = decoder.Decode(&feed); err != nil {
		return
	}

	ch := &common.Channel{
		Title:       feed.Channel.Title,
		Description: feed.Channel.Description,
		Links: []common.Link{
			common.Link{Href: feed.Channel.Link},
		},
	}

	if feed.Image != nil {
		ch.Image = common.Image{
			Title: feed.Image.Title,
			Url:   feed.Image.URL,
			Link:  feed.Image.Link,
		}
	}

	for _, item := range feed.Items {
		i := &common.Item{
			Title: item.Title,
			Links: []common.Link{
				common.Link{Href: item.Link},
			},
			Author:  common.Author{Name: item.DcCreator},
			PubDate: item.DcDate,
			Categories: []common.Category{
				common.Category{Domain: "", Text: item.DcSubject},
			},
			Content: &common.Content{Text: item.ContentEncoded},
		}

		ch.Items = append(ch.Items, i)
	}

	foundChannels = append(foundChannels, ch)

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

	dcModule
	syModule
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

	dcModule
}

type rdfItem struct {
	About       string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`

	dcModule
	contentModule
}

// http://web.resource.org/rss/1.0/modules/dc/
type dcModule struct {
	DcTitle       string `xml:"http://purl.org/dc/elements/1.1/ title"`
	DcCreator     string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	DcSubject     string `xml:"http://purl.org/dc/elements/1.1/ subject"`
	DcDescription string `xml:"http://purl.org/dc/elements/1.1/ description"`
	DcPublisher   string `xml:"http://purl.org/dc/elements/1.1/ publisher"`
	DcContributor string `xml:"http://purl.org/dc/elements/1.1/ contributor"`
	DcDate        string `xml:"http://purl.org/dc/elements/1.1/ date"`
	DcType        string `xml:"http://purl.org/dc/elements/1.1/ type"`
	DcFormat      string `xml:"http://purl.org/dc/elements/1.1/ format"`
	DcIdentifier  string `xml:"http://purl.org/dc/elements/1.1/ identifier"`
	DcSource      string `xml:"http://purl.org/dc/elements/1.1/ source"`
	DcLanguage    string `xml:"http://purl.org/dc/elements/1.1/ language"`
	DcRelation    string `xml:"http://purl.org/dc/elements/1.1/ relation"`
	DcCoverage    string `xml:"http://purl.org/dc/elements/1.1/ coverage"`
	DcRights      string `xml:"http://purl.org/dc/elements/1.1/ rights"`
}

// http://web.resource.org/rss/1.0/modules/syndication/
type syModule struct {
	// 'hourly' | 'daily' | 'weekly' | 'monthly' | 'yearly'
	SyUpdatePeriod    string    `xml:"http://purl.org/rss/1.0/modules/syndication/ updatePeriod"`
	SyUpdateFrequency uint      `xml:"http://purl.org/rss/1.0/modules/syndication/ updateFrequency"`
	SyUpdateBase      time.Time `xml:"http://purl.org/rss/1.0/modules/syndication/ updateBase"`
}

// http://web.resource.org/rss/1.0/modules/content/
type contentModule struct {
	ContentEncoded string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}
