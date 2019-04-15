package rss

import (
	"os"
	"testing"

	"hawx.me/code/assert"
)

func TestAuthor(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("../testdata/boing.rss")
	defer file.Close()

	channels, err := new(Parser).Read(file, nil)
	assert.Nil(err)

	if assert.Len(channels, 1) {
		channel := channels[0]

		if assert.Len(channel.Items, 25) {
			item := channel.Items[0]

			assert.Equal("Cory Doctorow", item.Author.Name)
		}
	}
}

func TestMediaExtensions(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("../testdata/media_extensions.rss")
	defer file.Close()

	channels, err := new(Parser).Read(file, nil)
	assert.Nil(err)

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("Media Extensions Testcase", channel.Title)

		if assert.Len(channel.Items, 2) {
			items := channel.Items
			assert.Equal("1", items[0].Title)
			assert.Equal("http://example.com/images/1.jpg", items[0].Thumbnail.URL)

			assert.Equal("2", items[1].Title)
			assert.Equal("http://example.com/images/2.jpg", items[1].Thumbnail.URL)
			assert.Equal(100, items[1].Thumbnail.Width)
			assert.Equal(123, items[1].Thumbnail.Height)
		}
	}
}
