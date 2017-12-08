package mapping

import (
	"html"
	"log"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"hawx.me/code/riviera/feed/common"
	"hawx.me/code/riviera/river/riverjs"
)

// DefaultMapping will always return an item. It: attempts to parse the PubDate,
// otherwise uses the current time; truncates the description to 280 characters;
// finds the correct Link and PermaLink; copies any Enclosures; and fills out
// the other properties by copying the correct values.
func DefaultMapping(item *common.Item) *riverjs.Item {
	pubDate, err := item.ParsedPubDate()
	if err != nil {
		log.Println("DefaultMapping/time:", err)
		pubDate = time.Now()
	}

	i := &riverjs.Item{
		Body:       stripAndCrop(item.Description),
		PubDate:    riverjs.RssTime{pubDate},
		Title:      html.UnescapeString(item.Title),
		Id:         item.Key(),
		Comments:   item.Comments,
		Enclosures: []riverjs.Enclosure{},
	}

	if item.Guid != nil && item.Guid.IsPermaLink {
		i.PermaLink = item.Guid.Guid
		i.Link = item.Guid.Guid
	}

	if len(item.Links) > 0 {
		i.PermaLink = item.Links[0].Href
		i.Link = item.Links[0].Href

		for _, link := range item.Links {
			if link.Rel == "alternate" {
				i.PermaLink = link.Href
				i.Link = link.Href
			}

			if link.Rel == "enclosure" {
				i.Enclosures = append(i.Enclosures, riverjs.Enclosure{
					Url:  link.Href,
					Type: link.Type,
				})
			}
		}
	}

	if item.Content != nil {
		i.Body = stripAndCrop(item.Content.Text)
	}

	for _, enclosure := range item.Enclosures {
		i.Enclosures = append(i.Enclosures, riverjs.Enclosure{
			Url:    enclosure.Url,
			Type:   enclosure.Type,
			Length: enclosure.Length,
		})
	}

	if item.Thumbnail != nil {
		i.Thumbnail = &riverjs.Thumbnail{
			Url: item.Thumbnail.Url,
		}

		if item.Thumbnail.Height > 0 {
			i.Thumbnail.Height = &item.Thumbnail.Height
		}

		if item.Thumbnail.Width > 0 {
			i.Thumbnail.Width = &item.Thumbnail.Width
		}
	}

	return i
}

// Strips html markup, then limits to 280 characters. If the original text was
// longer than 280 chars, an ellipsis is appended.
func stripAndCrop(content string) string {
	content = processString(content,
		strings.NewReplacer("\n", " ").Replace,
		strings.NewReplacer("  ", " ").Replace,
		strings.TrimSpace,
		html.UnescapeString,
		html.UnescapeString,
		sanitize.HTML)

	if len(content) <= 280 {
		return content
	}

	return strings.TrimSpace(content[0:279]) + "…"
}

func processString(in string, fs ...func(string) string) string {
	for _, f := range fs {
		in = f(in)
	}

	return in
}
