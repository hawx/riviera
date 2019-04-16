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
			assert.Equal("<p>This is a blog post.</p><p>With paragraphs.</p><p>As you might expect.</p>", channel.Items[0].Content.Text)
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
			assert.Equal("Despite not checking FB (always a good start), I’ve found it takes a while (at least a day?) to stop feeling compelled to check, keep up with, or “work on” whatever projects (job related or not) are top of mind. Email, Slack, etc.", item.Title)
			assert.Equal("Despite not checking FB (always a good start), I’ve found it takes a while (at least a day?) to stop feeling compelled to check, keep up with, or “work on” whatever projects (job related or not) are top of mind. Email, Slack, etc.", item.Content.Text)
			assert.Equal("https://tantek.com/2019/093/t2/stop-check-email-slack", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-04-03 23:49-0700", item.PubDate)

			item = channel.Items[1]
			assert.Equal("https://tantek.com/2019/093/t1/few-days-off", item.GUID.GUID)
			assert.Equal("Taking a few days off from many things for warmer, sunnier shores, and a yoga retreat.Still haven’t checked Facebook notifications since the last retreat, January over a year ago.Brought a notepad & books. Going to read, write, post what surfaces.", item.Title)
			assert.Equal("Taking a few days off from many things for warmer, sunnier shores, and a yoga retreat.<br class=\"auto-break\"><br class=\"auto-break\">Still haven’t checked Facebook notifications since the last retreat, January over a year ago.<br class=\"auto-break\"><br class=\"auto-break\">Brought a notepad &amp; books. Going to read, write, post what surfaces.", item.Content.Text)
			assert.Equal("https://tantek.com/2019/093/t1/few-days-off", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-04-03 07:53-0700", item.PubDate)

			item = channel.Items[2]
			assert.Equal("https://tantek.com/2019/090/t4/anniversaries-cruel-intentions-web30-heathers", item.GUID.GUID)
			assert.Equal("More anniversaries (this month) in addition to #TheMatrix:20th: #CruelIntentions https://hellogiggles.com/reviews-coverage/movies/why-kathryn-cruel-intentions-still-matters-2019/30th: #Web30 #TimBL information management proposal http://info.cern.ch/Proposal.html30 years ago today: #Heathers movie release https://www.newyorker.com/culture/touchstones/an-appreciation-of-the-dark-comedy-heathers", item.Title)
			assert.Equal("More anniversaries (this month) in addition to #<span class=\"p-category auto-tag\">TheMatrix:</span><br class=\"auto-break\"><br class=\"auto-break\">20th: #<span class=\"p-category auto-tag\">CruelIntentions</span> <a class=\"auto-link\" href=\"https://hellogiggles.com/reviews-coverage/movies/why-kathryn-cruel-intentions-still-matters-2019/\">https://hellogiggles.com/reviews-coverage/movies/why-kathryn-cruel-intentions-still-matters-2019/</a><br class=\"auto-break\"><br class=\"auto-break\">30th: #<span class=\"p-category auto-tag\">Web30</span> #<span class=\"p-category auto-tag\">TimBL</span> information management proposal <a class=\"auto-link\" href=\"http://info.cern.ch/Proposal.html\">http://info.cern.ch/Proposal.html</a><br class=\"auto-break\"><br class=\"auto-break\">30 years ago today: #<span class=\"p-category auto-tag\">Heathers</span> movie release <a class=\"auto-link\" href=\"https://www.newyorker.com/culture/touchstones/an-appreciation-of-the-dark-comedy-heathers\">https://www.newyorker.com/culture/touchstones/an-appreciation-of-the-dark-comedy-heathers</a>", item.Content.Text)
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
			assert.Equal("<p>\nThe webmention.io /api endpoints accept a 'target' query parameter which currently must be an absolute URL. This proposal extends the 'target' query parameter to also accept a protocol relative URL (i.e. starting with &#34;//&#34;) and return all mentions of that target with any protocol (e.g. both 'http:' and 'https:').\n</p>\n<p>\nUse-case: this will allow sites which have migrated from http to https, or which still serve both http and https and accept webmentions for both, to easily query webmention.io for all mentions to either http/https versions of their permalinks, to show webmentions regardless of which protocol was used in their webmention target URLs.\n</p>\n<p>\nTest-case: http://tantek.com/2019/065/e1/homebrew-website-club-sf is live with an iframe that embeds a display of RSVP webmentions via a service using a protocol relative URL in the target param to a webmention.io api call. Check both of these:\n</p>\n<ul>\n<li><a href=\"http://tantek.com/2019/065/e1/homebrew-website-club-sf\">http://tantek.com/2019/065/e1/homebrew-website-club-sf</a></li>\n<li><a href=\"https://tantek.com/2019/065/e1/homebrew-website-club-sf\">https://tantek.com/2019/065/e1/homebrew-website-club-sf</a></li>\n</ul>\n<p>\nAnd you should see at least one RSVP displayed in the bottom from v2.jacky.wtf, likely with a green checkmark ✅. Here’s a direct link to just the RSVP display for that post but hardcoded to 'http:' mentions only: <a href=\"https://stream.thatmustbe.us/?url=https%3A%2F%2Fwebmention.io%2Fapi%2Fmentions.jf2%3Fwm-property%3Drsvp%26sort-by%3Drsvp%26target%3Dhttp%3A%2F%2Ftantek.com%2F2019%2F065%2Fe1%2Fhomebrew-website-club-sf&amp;op=jf2-mf2&amp;ashtml=1&amp;style=img%7Bheight:44px;vertical-align:bottom%7D.h-feed%3E.p-name,.h-entry%3E.u-url:first-child,.h-entry%3E.p-name,.u-url%3E.p-name,.p-in-reply-to,.p-like-of,.p-mention-of,*%5Bclass%5E=p-wm-%5D,.dt-published,.p-syndication%7Bdisplay:none%7D.h-card%7Bfloat:left;border:solid%201px%20%23999;margin:0%202px%202px%200;line-height:44px%7D.p-rsvp:before%7Bcontent:%22%E2%98%85%22;float:left;color:white;background:blue;font-size:10px;line-height:1.5;width:13px;height:13px;text-align:center;margin:35px%200%20-3px%20-14px;border-radius:3px;position:relative;z-index:1%7D.p-rsvp%5Bvalue=yes%5D:before%7Bcontent:%22%E2%9C%85%22;font-size:13px;margin:34px%200%20-3px%20-16px;background:none%7D.p-rsvp%5Bvalue=no%5D:before%7Bcontent:%22%E2%9D%8C%22;background:white;color:red;font-size:13px%7Dbody%7Bmargin:0%7Dimg%5Bsrc=%22%22%5D%7Bwidth:0.6em%7D\">v2.jacky.wtf RSVP</a>.\n</p>", item.Content.Text)
			assert.Equal("https://tantek.com/2019/065/b1/webmention-io-target-accept-protocol-relative-url", item.Links[0].Href)
			assert.Equal("alternate", item.Links[0].Rel)
			assert.Equal("2019-03-06 10:29-0800", item.PubDate)
		}
	}
}
