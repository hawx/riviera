package rss

// http://www.rssboard.org/rss-specification

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"

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
			if t.Name.Space == "" && t.Name.Local == "rss" {
				for _, attr := range t.Attr {
					if attr.Name.Space == "" && attr.Name.Local == "version" {
						p := strings.Index(attr.Value, ".")
						major, _ := strconv.Atoi(attr.Value[0:p])
						minor, _ := strconv.Atoi(attr.Value[p+1 : len(attr.Value)])

						return !(major > 2 || (major == 2 && minor > 0))
					}
				}
			}

			return false
		}
	}

	return false
}

func (Parser) Read(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) (foundChannels []*data.Channel, err error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset

	var feed rssFeed
	if err = decoder.Decode(&feed); err != nil {
		return
	}

	ch := &data.Channel{
		Title:          feed.Channel.Title,
		Description:    feed.Channel.Description,
		Language:       feed.Channel.Language,
		Copyright:      feed.Channel.Copyright,
		ManagingEditor: feed.Channel.ManagingEditor,
		WebMaster:      feed.Channel.WebMaster,
		PubDate:        feed.Channel.PubDate,
		LastBuildDate:  feed.Channel.LastBuildDate,
		Docs:           feed.Channel.Docs,
		TTL:            feed.Channel.TTL,
		Rating:         feed.Channel.Rating,
	}

	for _, link := range feed.Channel.Links {
		if link.XMLName.Space == "http://www.w3.org/2005/Atom" {
			ch.Links = append(ch.Links, data.Link{
				Href:     link.Href,
				Rel:      link.Rel,
				Type:     link.Type,
				HrefLang: link.HrefLang,
			})
		} else {
			ch.Links = append(ch.Links, data.Link{
				Href: link.Text,
			})
		}
	}

	for _, category := range feed.Channel.Category {
		ch.Categories = append(ch.Categories, data.Category{
			Domain: category.Domain,
			Text:   category.Text,
		})
	}

	if feed.Channel.Generator != nil {
		ch.Generator = data.Generator{
			Text: *feed.Channel.Generator,
		}
	}

	if feed.Channel.SkipHours != nil {
		for _, hour := range feed.Channel.SkipHours.Hours {
			ch.SkipHours = append(ch.SkipHours, hour)
		}
	}

	if feed.Channel.SkipDays != nil {
		for _, day := range feed.Channel.SkipDays.Days {
			ch.SkipDays = append(ch.SkipDays, days[day])
		}
	}

	if feed.Channel.Image != nil {
		ch.Image = data.Image{
			Title:       feed.Channel.Image.Title,
			Url:         feed.Channel.Image.URL,
			Link:        feed.Channel.Image.Link,
			Width:       feed.Channel.Image.Width,
			Height:      feed.Channel.Image.Height,
			Description: feed.Channel.Image.Description,
		}
	}

	if feed.Channel.Cloud != nil {
		ch.Cloud = data.Cloud{
			Domain:            feed.Channel.Cloud.Domain,
			Port:              feed.Channel.Cloud.Port,
			Path:              feed.Channel.Cloud.Path,
			RegisterProcedure: feed.Channel.Cloud.RegisterProcedure,
			Protocol:          feed.Channel.Cloud.Protocol,
		}
	}

	for _, item := range feed.Channel.Items {
		i := &data.Item{
			Title:       item.Title,
			Description: strings.TrimSpace(item.Description),
			Comments:    item.Comments,
			PubDate:     item.PubDate,
		}

		for _, link := range item.Links {
			if link.XMLName.Space == "http://www.w3.org/2005/Atom" {
				i.Links = append(i.Links, data.Link{
					Href:     link.Href,
					Rel:      link.Rel,
					Type:     link.Type,
					HrefLang: link.HrefLang,
				})
			} else {
				i.Links = append(i.Links, data.Link{
					Href: link.Text,
				})
			}
		}

		if item.Author != nil {
			i.Author.Name = *item.Author
		} else if item.Creator != nil {
			i.Author.Name = *item.Creator
		}

		if item.Guid != nil {
			i.Guid = &data.Guid{
				Guid:        item.Guid.Text,
				IsPermaLink: item.Guid.IsPermaLink == "true",
			}
		}

		for _, category := range item.Category {
			i.Categories = append(i.Categories, data.Category{
				Domain: category.Domain,
				Text:   category.Text,
			})
		}

		for _, enclosure := range item.Enclosure {
			i.Enclosures = append(i.Enclosures, data.Enclosure{
				Url:    enclosure.URL,
				Length: enclosure.Length,
				Type:   enclosure.Type,
			})
		}

		if item.Source != nil {
			i.Source = &data.Source{
				Url:  item.Source.URL,
				Text: item.Source.Text,
			}
		}

		ch.Items = append(ch.Items, i)
	}

	foundChannels = append(foundChannels, ch)

	return
}

// Commentary taken from http://www.rssboard.org/rss-specification

// At the top level, a RSS document is a <rss> element, with a mandatory
// attribute called version, that specifies the version of RSS that the document
// conforms to. If it conforms to this specification, the version attribute must
// be 2.0.
type rssFeed struct {
	XMLName xml.Name `xml:"rss"`

	// Subordinate to the <rss> element is a single <channel> element, which
	// contains information about the channel (metadata) and its contents.
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`

	// required elements
	Title       string    `xml:"title"`
	Links       []rssLink `xml:"link"`
	Description string    `xml:"description"`

	// optional elements
	Language       string        `xml:"language"`
	Copyright      string        `xml:"copyright"`
	ManagingEditor string        `xml:"managingEditor"`
	WebMaster      string        `xml:"webMaster"`
	PubDate        string        `xml:"pubDate"`
	LastBuildDate  string        `xml:"lastBuildDate"`
	Category       []rssCategory `xml:"category"`
	Generator      *string       `xml:"generator"`
	Docs           string        `xml:"docs"`
	Cloud          *rssCloud     `xml:"cloud"`
	TTL            int           `xml:"ttl"`
	Image          *rssImage     `xml:"image"`
	Rating         string        `xml:"rating"`
	SkipHours      *rssSkipHours `xml:"skipHours"`
	SkipDays       *rssSkipDays  `xml:"skipDays"`
}

type rssLink struct {
	XMLName xml.Name

	// rss value
	Text string `xml:",chardata"`

	// atom attributes
	Href     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr"`
	Type     string `xml:"type,attr"`
	HrefLang string `xml:"hreflang,attr"`
}

type rssSkipHours struct {
	Hours []int `xml:"hour"`
}

type rssSkipDays struct {
	Days []string `xml:"day"`
}

type rssImage struct {
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Width       int    `xml:"width"`
	Height      int    `xml:"height"`
	Description string `xml:"description"`
}

type rssCloud struct {
	Domain            string `xml:"domain,attr"`
	Port              int    `xml:"port,attr"`
	Path              string `xml:"path,attr"`
	RegisterProcedure string `xml:"registerProcedure,attr"`
	Protocol          string `xml:"protocol,attr"`
}

type rssItem struct {
	Title       string         `xml:"title"`
	Links       []rssLink      `xml:"link"`
	Description string         `xml:"description"`
	Author      *string        `xml:"author"`
	Creator     *string        `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Category    []rssCategory  `xml:"category"`
	Comments    string         `xml:"comments"`
	Enclosure   []rssEnclosure `xml:"enclosure"`
	Guid        *rssGuid       `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	Source      *rssSource     `xml:"source"`
}

type rssCategory struct {
	Domain string `xml:"domain,attr"`
	Text   string `xml:",chardata"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type rssGuid struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Text        string `xml:",chardata"`
}

type rssSource struct {
	URL  string `xml:"url,attr"`
	Text string `xml:",chardata"`
}
