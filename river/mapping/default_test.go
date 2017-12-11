package mapping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/feed/common"
	"hawx.me/code/riviera/river/riverjs"
)

func TestDefaultMapping(t *testing.T) {
	fifty := 50
	oneHundred := 100

	testcases := []struct {
		name       string
		feedItem   *common.Item
		modelsItem *riverjs.Item
	}{
		{
			"standard",
			&common.Item{
				Title: "cool feed thang",
				Links: []common.Link{
					{Href: "http://example.com/now"},
					{Href: "http://example.org/this", Rel: "alternate"},
					{Href: "http://example.com/what"},
				},
				Description: "this is tha content",
				PubDate:     "Mon, 02 Jan 2006 20:04:19 UTC",
			},
			&riverjs.Item{
				PermaLink:  "http://example.org/this",
				Link:       "http://example.org/this",
				Body:       "this is tha content",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Title:      "cool feed thang",
				Id:         "cool feed thangMon, 02 Jan 2006 20:04:19 UTC",
				Comments:   "",
				Enclosures: []riverjs.Enclosure{},
			},
		},

		// Description
		{
			"description truncated",
			&common.Item{
				Title:       "cool feed thang",
				Description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit. Donec et mollis dolor. Praesent et diam eget libero egestas mattis sit amet vitae augue. Nam tincidunt congue enim, ut porta lorem lacinia consectetur. Donec ut libero sed arcu vehicula ultricies a non tortor. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean ut gravida lorem. Ut turpis felis, pulvinar a semper sed, adipiscing id dolor. Pellentesque auctor nisi id magna consequat sagittis. Curabitur dapibus enim sit amet elit pharetra tincidunt feugiat nisl imperdiet. Ut convallis libero in urna ultrices accumsan. Donec sed odio eros. Donec viverra mi quis quam pulvinar at malesuada arcu rhoncus. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. In rutrum accumsan ultricies. Mauris vitae nisi at sem facilisis semper ac in est.",
				PubDate:     "Mon, 02 Jan 2006 20:04:19 UTC",
			},
			&riverjs.Item{
				Body:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit. Donec et mollis dolor. Praesent et diam eget libero egestas mattis sit amet vitae augue. Nam tincidunt…",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Title:      "cool feed thang",
				Id:         "cool feed thangMon, 02 Jan 2006 20:04:19 UTC",
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"description unescaped",
			&common.Item{
				Description: "&apos;",
				PubDate:     "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:          "5",
			},
			&riverjs.Item{
				Body:       "'",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []riverjs.Enclosure{},
			},
		},

		// Title
		{
			"title unescaped",
			&common.Item{
				Title:   "&#8220;The purpose of the IoT is to give humans superpowers&#8221;",
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
			},
			&riverjs.Item{
				Title:      `“The purpose of the IoT is to give humans superpowers”`,
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []riverjs.Enclosure{},
			},
		},

		// Pubdate
		{
			"pubdate",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "-",
			},
			&riverjs.Item{
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "-",
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"pubdate in other format", // am I going to do all of these?
			&common.Item{
				PubDate: "2006-01-02T20:04:19+00:00",
				Id:      "-",
			},
			&riverjs.Item{
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.Local)},
				Id:         "-",
				Enclosures: []riverjs.Enclosure{},
			},
		},

		// Id
		{
			"id from id",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
			},
			&riverjs.Item{
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"id from guid",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &common.Guid{Guid: "200823-4545345-435543-45"},
			},
			&riverjs.Item{
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "200823-4545345-435543-45",
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"id from title and pubdate",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Title:   "hey",
			},
			&riverjs.Item{
				Title:      "hey",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "heyMon, 02 Jan 2006 20:04:19 UTC",
				Enclosures: []riverjs.Enclosure{},
			},
		},

		// PermaLink and Link
		{
			"links from guid",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &common.Guid{Guid: "5", IsPermaLink: true},
			},
			&riverjs.Item{
				Link:       "5",
				PermaLink:  "5",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"links from (first) links",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &common.Guid{Guid: "5", IsPermaLink: true},
				Links: []common.Link{
					{Href: "cool"},
					{Href: "ignored"},
				},
			},
			&riverjs.Item{
				Link:       "cool",
				PermaLink:  "cool",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"links from alternate links",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Guid:    &common.Guid{Guid: "5", IsPermaLink: true},
				Links: []common.Link{
					{Href: "cool"},
					{Href: "alt", Rel: "alternate"},
				},
			},
			&riverjs.Item{
				Link:       "alt",
				PermaLink:  "alt",
				PubDate:    riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:         "5",
				Enclosures: []riverjs.Enclosure{},
			},
		},

		// Enclosure
		{
			"enclosure",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
				Links: []common.Link{
					{Href: "what"},
					{Href: "thing", Type: "media/what", Rel: "enclosure"},
					{Href: "otherthing", Type: "media/what", Rel: "enclosure"},
				},
			},
			&riverjs.Item{
				Link:      "what",
				PermaLink: "what",
				PubDate:   riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:        "5",
				Enclosures: []riverjs.Enclosure{
					{Url: "thing", Type: "media/what"},
					{Url: "otherthing", Type: "media/what"},
				},
			},
		},

		// Thumbnail
		{
			"thumbnail",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
				Thumbnail: &common.Image{
					Url: "http://example.com/thumb",
				},
			},
			&riverjs.Item{
				PubDate: riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:      "5",
				Thumbnail: &riverjs.Thumbnail{
					Url: "http://example.com/thumb",
				},
				Enclosures: []riverjs.Enclosure{},
			},
		},
		{
			"thumbnail with size",
			&common.Item{
				PubDate: "Mon, 02 Jan 2006 20:04:19 UTC",
				Id:      "5",
				Thumbnail: &common.Image{
					Url:    "http://example.com/thumb",
					Height: 50,
					Width:  100,
				},
			},
			&riverjs.Item{
				PubDate: riverjs.RssTime{time.Date(2006, 1, 2, 20, 4, 19, 0, time.UTC)},
				Id:      "5",
				Thumbnail: &riverjs.Thumbnail{
					Url:    "http://example.com/thumb",
					Height: &fifty,
					Width:  &oneHundred,
				},
				Enclosures: []riverjs.Enclosure{},
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

func stringOfLength(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "X"
	}
	return s
}

func TestStripAndCrop(t *testing.T) {
	tcs := []struct {
		In, Out string
	}{
		{``, ``},
		{`Hey  what`, `Hey what`},
		{stringOfLength(280), stringOfLength(280)},
		{stringOfLength(281), stringOfLength(281)[0:279] + "…"},
		{stringOfLength(279) + "  ", stringOfLength(279)},
		{`&amp;`, `&amp;`},
		{`<p>`, ``},
		{`&lt;p&gt;`, ``},
		{`&amp;lt;p&amp;gt;`, ``},
		{`<p>Hello

there <a href="coolcat.jpg">pictur</a></p>


`, `Hello there pictur
`},
	}

	for _, tc := range tcs {
		assert.Equal(t, tc.Out, stripAndCrop(tc.In))
	}
}
