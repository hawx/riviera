package jsonfeed

import (
	"os"
	"testing"

	"hawx.me/code/assert"
)

func TestSimple(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/simple.json")
	defer file.Close()

	channels, err := new(Parser).Read(file, nil)
	assert.Nil(err)

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("My Example Feed", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("https://example.org/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("https://example.org/feed.json", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}

		if assert.Len(channel.Items, 2) {
			assert.Equal("2", channel.Items[0].Guid.Guid)
			assert.Equal("https://example.org/second-item", channel.Items[0].Links[0].Href)
			assert.Equal("alternate", channel.Items[0].Links[0].Rel)
			assert.Equal("This is a second item.", channel.Items[0].Content.Text)

			assert.Equal("1", channel.Items[1].Guid.Guid)
			assert.Equal("https://example.org/initial-post", channel.Items[1].Links[0].Href)
			assert.Equal("alternate", channel.Items[1].Links[0].Rel)
			assert.Equal("<p>Hello, world!</p>", channel.Items[1].Content.Text)

			if assert.Len(channel.Items[1].Categories, 2) {
				assert.Equal("test", channel.Items[1].Categories[0].Text)
				assert.Equal("other", channel.Items[1].Categories[1].Text)
			}
		}
	}
}

func TestJsonfeed(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/jsonfeed.json")
	defer file.Close()

	channels, err := new(Parser).Read(file, nil)
	assert.Nil(err)

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("JSON Feed", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("https://jsonfeed.org/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("https://jsonfeed.org/feed.json", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}
		assert.Equal("Brent Simmons and Manton Reece", channel.Author.Name)

		if assert.Len(channel.Items, 1) {
			assert.Equal("https://jsonfeed.org/2017/05/17/announcing_json_feed", channel.Items[0].Guid.Guid)
			assert.Equal("https://jsonfeed.org/2017/05/17/announcing_json_feed", channel.Items[0].Links[0].Href)
			assert.Equal("alternate", channel.Items[0].Links[0].Rel)
			assert.Equal("2017-05-17T08:02:12-07:00", channel.Items[0].PubDate)
			assert.Equal(`<p>We — Manton Reece and Brent Simmons — have noticed that JSON has become the developers’ choice for APIs, and that developers will often go out of their way to avoid XML. JSON is simpler to read and write, and it’s less prone to bugs.</p>

<p>So we developed JSON Feed, a format similar to <a href="http://cyber.harvard.edu/rss/rss.html">RSS</a> and <a href="https://tools.ietf.org/html/rfc4287">Atom</a> but in JSON. It reflects the lessons learned from our years of work reading and publishing feeds.</p>

<p><a href="https://jsonfeed.org/version/1">See the spec</a>. It’s at version 1, which may be the only version ever needed. If future versions are needed, version 1 feeds will still be valid feeds.</p>

<h4>Notes</h4>

<p>We have a <a href="https://github.com/manton/jsonfeed-wp">WordPress plugin</a> and, coming soon, a JSON Feed Parser for Swift. As more code is written, by us and others, we’ll update the <a href="https://jsonfeed.org/code">code</a> page.</p>

<p>See <a href="https://jsonfeed.org/mappingrssandatom">Mapping RSS and Atom to JSON Feed</a> for more on the similarities between the formats.</p>

<p>This website — the Markdown files and supporting resources — <a href="https://github.com/brentsimmons/JSONFeed">is up on GitHub</a>, and you’re welcome to comment there.</p>

<p>This website is also a blog, and you can subscribe to the <a href="https://jsonfeed.org/xml/rss.xml">RSS feed</a> or the <a href="https://jsonfeed.org/feed.json">JSON feed</a> (if your reader supports it).</p>

<p>We worked with a number of people on this over the course of several months. We list them, and thank them, at the bottom of the <a href="https://jsonfeed.org/version/1">spec</a>. But — most importantly — <a href="http://furbo.org/">Craig Hockenberry</a> spent a little time making it look pretty. :)</p>`, channel.Items[0].Content.Text)
		}
	}
}

func TestPodcast(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/podcast.json")
	defer file.Close()

	channels, err := new(Parser).Read(file, nil)
	assert.Nil(err)

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("The Record", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("http://therecord.co/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("http://therecord.co/feed.json", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}

		if assert.Len(channel.Items, 1) {
			item := channel.Items[0]
			assert.Equal("Special #1 - Chris Parrish", item.Title)
			assert.Equal("http://therecord.co/chris-parrish", item.Guid.Guid)
			assert.Equal("http://therecord.co/chris-parrish", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("Brent interviews Chris Parrish, co-host of The Record and one-half of Aged & Distilled.", item.Content.Text)
			assert.Equal("2014-05-09T14:04:00-07:00", item.PubDate)

			if assert.Len(item.Enclosures, 1) {
				enclosure := item.Enclosures[0]
				assert.Equal("http://therecord.co/downloads/The-Record-sp1e1-ChrisParrish.m4a", enclosure.Url)
				assert.Equal("audio/x-m4a", enclosure.Type)
				assert.Equal(int64(89970236), enclosure.Length)
			}
		}
	}
}
