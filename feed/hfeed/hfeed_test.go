package hfeed

import (
	"os"
	"testing"

	"hawx.me/code/assert"
)

func TestSimple(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/simple.html")
	defer file.Close()

	if ok := new(Parser).CanRead(file, nil); !assert.True(ok) {
		return
	}

	if _, err := file.Seek(0, 0); !assert.Nil(err) {
		return
	}

	channels, err := new(Parser).Read(file, nil)
	if !assert.Nil(err) {
		return
	}

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("My Blog", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("https://example.org/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("https://example.org/feed.json", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}

		if assert.Len(channel.Items, 1) {
			assert.Equal("2019/01/01/an-article", channel.Items[0].GUID.GUID)
			assert.Equal("An article", channel.Items[0].Title)
			assert.Equal("<p>This is a blog post.</p><p>With paragraphs.</p><p>As you might expect.</p>", channel.Items[0].Content.Text)
			assert.Equal("https://example.org/2019/01/01/an-article", channel.Items[0].Links[0].Href)
			assert.Equal("alternate", channel.Items[0].Links[0].Rel)
			assert.Equal("2019-01-01T12:00:00Z", channel.Items[0].PubDate)
		}
	}
}
