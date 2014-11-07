package river

import (
	"github.com/hawx/go-pkg-rss"
	"github.com/hawx/riviera/river/models"
	"github.com/kennygrant/sanitize"
)

func convertItem(item *feeder.Item) *models.Item {
	pubDate, err := item.ParsedPubDate()
	if err != nil {
		return nil
	}

	i := &models.Item{
		Body:    stripAndCrop(item.Description),
		PubDate: models.RssTime{pubDate},
		Title:   item.Title,
		Id:      item.Key(),
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
