package hfeed

import (
	"net/url"
	"os"
	"testing"

	"hawx.me/code/assert"
)

func TestSimple(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/simple.html")
	defer file.Close()

	rootURL, _ := url.Parse("https://example.org/")

	parser := Parser{}

	if ok := parser.CanRead(file, nil); !assert.True(ok) {
		return
	}

	if _, err := file.Seek(0, 0); !assert.Nil(err) {
		return
	}

	channels, err := parser.Read(file, rootURL, nil)
	if !assert.Nil(err) {
		return
	}

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("My Blog", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("https://example.org/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("https://example.org/", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}

		if assert.Len(channel.Items, 1) {
			assert.Equal("https://example.org/2019/01/01/an-article", channel.Items[0].GUID.GUID)
			assert.Equal("An article", channel.Items[0].Title)
			assert.Equal("This is a blog post.With paragraphs.As you might expect.", channel.Items[0].Content.Text)
			assert.Equal("https://example.org/2019/01/01/an-article", channel.Items[0].Links[0].Href)
			assert.Equal("alternate", channel.Items[0].Links[0].Rel)
			assert.Equal("2019-01-01T12:00:00Z", channel.Items[0].PubDate)
		}
	}
}

func TestRealLife(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/tantek.com.html")
	defer file.Close()

	rootURL, _ := url.Parse("https://tantek.com/")

	parser := Parser{}

	if ok := parser.CanRead(file, nil); !assert.True(ok) {
		return
	}

	if _, err := file.Seek(0, 0); !assert.Nil(err) {
		return
	}

	channels, err := parser.Read(file, rootURL, nil)
	if !assert.Nil(err) {
		return
	}

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("https://tantek.com/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("https://tantek.com/", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}

		if assert.Len(channel.Items, 39) {
			item := channel.Items[0]
			assert.Equal("https://tantek.com/2019/093/t2/stop-check-email-slack", item.GUID.GUID)
			assert.Equal("", item.Title)
			assert.Equal("Despite not checking FB (always a good start), I’ve found it takes a while (at least a day?) to stop feeling compelled to check, keep up with, or “work on” whatever projects (job related or not) are top of mind. Email, Slack, etc.", item.Content.Text)
			assert.Equal("https://tantek.com/2019/093/t2/stop-check-email-slack", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-04-03 23:49-0700", item.PubDate)

			item = channel.Items[1]
			assert.Equal("https://tantek.com/2019/093/t1/few-days-off", item.GUID.GUID)
			assert.Equal("", item.Title)
			assert.Equal("Taking a few days off from many things for warmer, sunnier shores, and a yoga retreat.Still haven’t checked Facebook notifications since the last retreat, January over a year ago.Brought a notepad & books. Going to read, write, post what surfaces.", item.Content.Text)
			assert.Equal("https://tantek.com/2019/093/t1/few-days-off", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-04-03 07:53-0700", item.PubDate)

			item = channel.Items[2]
			assert.Equal("https://tantek.com/2019/090/t4/anniversaries-cruel-intentions-web30-heathers", item.GUID.GUID)
			assert.Equal("", item.Title)
			assert.Equal("More anniversaries (this month) in addition to #TheMatrix:20th: #CruelIntentions https://hellogiggles.com/reviews-coverage/movies/why-kathryn-cruel-intentions-still-matters-2019/30th: #Web30 #TimBL information management proposal http://info.cern.ch/Proposal.html30 years ago today: #Heathers movie release https://www.newyorker.com/culture/touchstones/an-appreciation-of-the-dark-comedy-heathers", item.Content.Text)
			assert.Equal("https://tantek.com/2019/090/t4/anniversaries-cruel-intentions-web30-heathers", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-03-31 23:59-0700", item.PubDate)

			item = channel.Items[20]
			assert.Nil(item.GUID)
			assert.Equal("likes @jgarber’s tweet", item.Title)
			assert.Nil(item.Content)
			assert.Equal("https://tantek.com/2019/072/f8", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-03-13T18:54-0700", item.PubDate)

			item = channel.Items[35]
			assert.Equal("https://tantek.com/2019/065/b1/webmention-io-target-accept-protocol-relative-url", item.GUID.GUID)
			assert.Equal("Webmention.io /api target= param should accept protocol relative URL to return http + https mentions", item.Title)
			assert.Equal("\nThe webmention.io /api endpoints accept a 'target' query parameter which currently must be an absolute URL. This proposal extends the 'target' query parameter to also accept a protocol relative URL (i.e. starting with \"//\") and return all mentions of that target with any protocol (e.g. both 'http:' and 'https:').\n\n\nUse-case: this will allow sites which have migrated from http to https, or which still serve both http and https and accept webmentions for both, to easily query webmention.io for all mentions to either http/https versions of their permalinks, to show webmentions regardless of which protocol was used in their webmention target URLs.\n\n\nTest-case: http://tantek.com/2019/065/e1/homebrew-website-club-sf is live with an iframe that embeds a display of RSVP webmentions via a service using a protocol relative URL in the target param to a webmention.io api call. Check both of these:\n\n\nhttp://tantek.com/2019/065/e1/homebrew-website-club-sf\nhttps://tantek.com/2019/065/e1/homebrew-website-club-sf\n\n\nAnd you should see at least one RSVP displayed in the bottom from v2.jacky.wtf, likely with a green checkmark ✅. Here’s a direct link to just the RSVP display for that post but hardcoded to 'http:' mentions only: v2.jacky.wtf RSVP.\n", item.Content.Text)
			assert.Equal("https://tantek.com/2019/065/b1/webmention-io-target-accept-protocol-relative-url", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-03-06 10:29-0800", item.PubDate)
		}
	}
}

func TestTopLevelEntries(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/me.hawx.me.html")
	defer file.Close()

	rootURL, _ := url.Parse("https://me.hawx.me/")

	parser := Parser{}

	if ok := parser.CanRead(file, nil); !assert.True(ok) {
		return
	}

	if _, err := file.Seek(0, 0); !assert.Nil(err) {
		return
	}

	channels, err := parser.Read(file, rootURL, nil)
	if !assert.Nil(err) {
		return
	}

	if assert.Len(channels, 1) {
		channel := channels[0]

		assert.Equal("", channel.Title)
		if assert.Len(channel.Links, 2) {
			assert.Equal("https://me.hawx.me/", channel.Links[0].Href)
			assert.Equal("alternate", channel.Links[0].Rel)

			assert.Equal("https://me.hawx.me/", channel.Links[1].Href)
			assert.Equal("self", channel.Links[1].Rel)
		}

		if assert.Len(channel.Items, 25) {
			item := channel.Items[0]
			assert.Nil(item.GUID)
			assert.Equal("@SimpsonsQOTD's tweet\n          at\n            \n              20:14", item.Title)
			assert.Nil(item.Content)
			assert.Equal("https://me.hawx.me/entry/3069f49f-f5d7-486b-b7e7-e47ac3528bfb", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2020-02-28T20:14:31Z", item.PubDate)

			item = channel.Items[1]
			assert.Nil(item.GUID)
			assert.Equal("@VeneficusIpse's tweet\n          at\n            \n              20:15", item.Title)
			assert.Nil(item.Content)
			assert.Equal("https://me.hawx.me/entry/5c386721-f474-4771-bd9b-80987ee8e18f", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2020-02-28T20:15:14Z", item.PubDate)

			item = channel.Items[2]
			assert.Nil(item.GUID)
			assert.Equal("", item.Title)
			assert.Equal("Blog update: \n- WebSub ✔️\n- Flickr ✔️, but only likes (https://me.hawx.me/entry/ba11f3aa-4c9b-4544-93d7-2fc74220cbda), and replies at the moment", item.Content.Text)
			assert.Equal("https://me.hawx.me/entry/a844f900-034b-4e6a-aa10-3146de45bf84", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2020-02-27T19:19:35Z", item.PubDate)

			item = channel.Items[12]
			assert.Nil(item.GUID)
			assert.Equal("", item.Title)
			assert.Equal("Can I reply to myself?", item.Content.Text)
			assert.Equal("https://me.hawx.me/entry/12020c6f-4fa7-40b3-a4bb-5d458ec650a1", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2020-02-13T19:43:17Z", item.PubDate)
		}
	}
}
