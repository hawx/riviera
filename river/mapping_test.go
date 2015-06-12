package river

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/models"
)

func TestDefaultMapping(t *testing.T) {
	testcases := []struct {
		name       string
		feedItem   *feed.Item
		modelsItem *models.Item
	}{
		{
			"standard",
			&feed.Item{
				Title: "cool feed thang",
				Links: []*feed.Link{
					{Href: "http://example.com/now"},
					{Href: "http://example.org/this", Rel: "alternate"},
					{Href: "http://example.com/what"},
				},
				Description: "this is tha content",
				PubDate:     "Mon, 02 Jan 2006 20:04:19 UTC",
			},
			&models.Item{
				PermaLink:  "http://example.org/this",
				Link:       "http://example.org/this",
				Body:       "this is tha content",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Title:      "cool feed thang",
				Id:         "cool feed thangMon, 02 Jan 2006 20:04:19 UTC",
				Comments:   "",
				Enclosures: []models.Enclosure{},
			},
		},

		// Description
		{
			"description truncated",
			&feed.Item{
				Title:       "cool feed thang",
				Description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit. Donec et mollis dolor. Praesent et diam eget libero egestas mattis sit amet vitae augue. Nam tincidunt congue enim, ut porta lorem lacinia consectetur. Donec ut libero sed arcu vehicula ultricies a non tortor. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean ut gravida lorem. Ut turpis felis, pulvinar a semper sed, adipiscing id dolor. Pellentesque auctor nisi id magna consequat sagittis. Curabitur dapibus enim sit amet elit pharetra tincidunt feugiat nisl imperdiet. Ut convallis libero in urna ultrices accumsan. Donec sed odio eros. Donec viverra mi quis quam pulvinar at malesuada arcu rhoncus. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. In rutrum accumsan ultricies. Mauris vitae nisi at sem facilisis semper ac in est.",
				PubDate:     "Mon, 02 Jan 2006 20:04:19 UTC",
			},
			&models.Item{
				Body:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit. Donec et mollis dolor. Praesent et diam eget libero egestas mattis sit amet vitae augue. Nam tincidunt c...",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Title:      "cool feed thang",
				Id:         "cool feed thangMon, 02 Jan 2006 20:04:19 UTC",
				Enclosures: []models.Enclosure{},
			},
		},
		{
			"description unescaped",
			&feed.Item{
				Description: "&apos;",
				PubDate:     "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:          "5",
			},
			&models.Item{
				Body:       "'",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []models.Enclosure{},
			},
		},

		// Pubdate
		{
			"pubdate",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "-",
			},
			&models.Item{
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "-",
				Enclosures: []models.Enclosure{},
			},
		},
		{
			"pubdate in other format", // am I going to do all of these?
			&feed.Item{
				PubDate: "2006-01-02T20:04:19+00:00",
				Id:      "-",
			},
			&models.Item{
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.Local)},
				Id:         "-",
				Enclosures: []models.Enclosure{},
			},
		},

		// Id
		{
			"id from id",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
			},
			&models.Item{
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []models.Enclosure{},
			},
		},
		{
			"id from guid",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &feed.Guid{Guid: "200823-4545345-435543-45"},
			},
			&models.Item{
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "200823-4545345-435543-45",
				Enclosures: []models.Enclosure{},
			},
		},
		{
			"id from title and pubdate",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Title:   "hey",
			},
			&models.Item{
				Title:      "hey",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "heyMon, 02 Jan 2006 20:04:19 UTC",
				Enclosures: []models.Enclosure{},
			},
		},

		// PermaLink and Link
		{
			"links from guid",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &feed.Guid{Guid: "5", IsPermaLink: true},
			},
			&models.Item{
				Link:       "5",
				PermaLink:  "5",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []models.Enclosure{},
			},
		},
		{
			"links from (first) links",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &feed.Guid{Guid: "5", IsPermaLink: true},
				Links: []*feed.Link{
					{Href: "cool"},
					{Href: "ignored"},
				},
			},
			&models.Item{
				Link:       "cool",
				PermaLink:  "cool",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []models.Enclosure{},
			},
		},
		{
			"links from alternate links",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &feed.Guid{Guid: "5", IsPermaLink: true},
				Links: []*feed.Link{
					{Href: "cool"},
					{Href: "alt", Rel: "alternate"},
				},
			},
			&models.Item{
				Link:       "alt",
				PermaLink:  "alt",
				PubDate:    models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []models.Enclosure{},
			},
		},

		// Enclosure
		{
			"enclosure",
			&feed.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
				Links: []*feed.Link{
					{Href: "what"},
					{Href: "thing", Type: "media/what", Rel: "enclosure"},
					{Href: "otherthing", Type: "media/what", Rel: "enclosure"},
				},
			},
			&models.Item{
				Link:      "what",
				PermaLink: "what",
				PubDate:   models.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:        "5",
				Enclosures: []models.Enclosure{
					{Url: "thing", Type: "media/what"},
					{Url: "otherthing", Type: "media/what"},
				},
			},
		},
	}

	assert := assert.New(t)

	for _, tc := range testcases {
		expected := tc.modelsItem
		mapped := DefaultMapping(tc.feedItem)

		assert.Equal(expected.Body, mapped.Body, tc.name)
		assert.Equal(expected.PermaLink, mapped.PermaLink, tc.name)
		assert.Equal(expected.PubDate, mapped.PubDate, tc.name)
		assert.Equal(expected.Title, mapped.Title, tc.name)
		assert.Equal(expected.Link, mapped.Link, tc.name)
		assert.Equal(expected.Id, mapped.Id, tc.name)
		assert.Equal(expected.Comments, mapped.Comments, tc.name)
		assert.Equal(expected.Enclosures, mapped.Enclosures, tc.name)
	}
}