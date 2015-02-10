package river

import (
	"github.com/hawx/riviera/feed"
	"github.com/hawx/riviera/river/models"
	"github.com/kennygrant/sanitize"

	"log"
	"time"
)

func convertItem(item *feed.Item) *models.Item {
	pubDate, err := item.ParsedPubDate()
	if err != nil {
		log.Println(err)
		pubDate = time.Now()
	}

	i := &models.Item{
		Body:       stripAndCrop(item.Description),
		PubDate:    models.RssTime{pubDate},
		Title:      item.Title,
		Id:         item.Key(),
		Comments:   item.Comments,
		Enclosures: []models.Enclosure{},
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
		}
	}

	if item.Content != nil {
		i.Body = stripAndCrop(item.Content.Text)
	}

	for _, enclosure := range item.Enclosures {
		i.Enclosures = append(i.Enclosures, models.Enclosure{
			Url:    enclosure.Url,
			Type:   enclosure.Type,
			Length: enclosure.Length,
		})
	}

	return i
}

// Strips html markup, then limits to 280 characters. If the original text was
// longer than 280 chars, three periods are appended.
func stripAndCrop(content string) string {
	content = sanitize.HTML(content)

	if len(content) < 280 {
		return content
	}

	return content[0:280] + "..."
}
