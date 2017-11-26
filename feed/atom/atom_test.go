package atom

import (
	"os"
	"testing"

	"hawx.me/code/assert"
)

func TestMediaExtensions(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("../testdata/media_extensions.atom")
	defer file.Close()

	channels, err := new(Parser).Read(file, nil)
	assert.Nil(err)

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("Media Extensions Testcase", channel.Title)

		if assert.Len(channel.Items, 2) {
			items := channel.Items
			assert.Equal("1", items[0].Title)
			assert.Equal("http://example.com/images/1.jpg", items[0].Thumbnail.Url)

			assert.Equal("2", items[1].Title)
			assert.Equal("http://example.com/images/2.jpg", items[1].Thumbnail.Url)
			assert.Equal(100, items[1].Thumbnail.Width)
			assert.Equal(123, items[1].Thumbnail.Height)
		}
	}
}
