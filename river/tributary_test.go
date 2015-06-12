package river

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/internal/persistence"
	"hawx.me/code/riviera/river/models"
)

func TestTributary(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Boing Boing</title>
    <link>http://boingboing.net</link>
    <description>Brain candy for Happy Mutants</description>
    <language>en-US</language>
    <lastBuildDate>Wed, 27 Mar 2013 12:30:18 WHAT</lastBuildDate>
    <item>
      <title>Save Noisebridge!</title>
      <link>http://feedproxy.google.com/~r/boingboing/iBag/~3/EKKb-61Ismc/story01.htm</link>
      <pubDate>Wed, 27 Mar 2013 12:40:18 UTC</pubDate>
      <guid isPermaLink="false">http://boingboing.net/?p=221544</guid>
      <description>A reader writes, "Noisebridge, San Francisco's Hackerspace, is having some hard times, so we're throwing an epic benefit and party this Saturday, to include eclectic performers, interactive art, a raffle and more! For more details, if any BBers want to put on demos or ideas share them.&lt;img width='1' height='1' src='http://rss.feedsportal.com/c/35208/f/653965/s/2a105a0e/mf.gif' border='0'/&gt;&lt;div class='mf-viral'&gt;&lt;table border='0'&gt;&lt;tr&gt;&lt;td valign='middle'&gt;&lt;a href="http://share.feedsportal.com/viral/sendEmail.cfm?lang=en&amp;title=Save+Noisebridge%21&amp;link=http%3A%2F%2Fboingboing.net%2F2013%2F03%2F27%2Fsave-noisebridge.html" target="_blank"&gt;&lt;img src="http://res3.feedsportal.com/images/emailthis2.gif" border="0" /&gt;&lt;/a&gt;&lt;/td&gt;&lt;td valign='middle'&gt;&lt;a href="http://res.feedsportal.com/viral/bookmark.cfm?title=Save+Noisebridge%21&amp;link=http%3A%2F%2Fboingboing.net%2F2013%2F03%2F27%2Fsave-noisebridge.html" target="_blank"&gt;&lt;img src="http://res3.feedsportal.com/images/bookmark.gif" border="0" /&gt;&lt;/a&gt;&lt;/td&gt;&lt;/tr&gt;&lt;/table&gt;&lt;/div&gt;</description>
    </item>
  </channel>
</rss>`))
	}))
	defer s.Close()

	db := memdata.Open()
	bucket, _ := persistence.NewBucket(db, "-")

	ch := make(chan models.Feed, 1)

	tributary := newTributary(bucket, s.URL, time.Minute, DefaultMapping)
	tributary.OnUpdate = func(f models.Feed) {
		ch <- f
	}

	expected := models.Feed{
		FeedUrl:         s.URL,
		WebsiteUrl:      "http://boingboing.net",
		FeedTitle:       "Boing Boing",
		FeedDescription: "Brain candy for Happy Mutants",
		WhenLastUpdate:  models.RssTime{time.Now()},
		Items: []models.Item{{
			Title:      "Save Noisebridge!",
			Link:       "http://feedproxy.google.com/~r/boingboing/iBag/~3/EKKb-61Ismc/story01.htm",
			PermaLink:  "http://feedproxy.google.com/~r/boingboing/iBag/~3/EKKb-61Ismc/story01.htm",
			Id:         "http://boingboing.net/?p=221544",
			PubDate:    models.RssTime{time.Date(2013, 03, 27, 12, 40, 18, 0, time.UTC)},
			Body:       "A reader writes, \"Noisebridge, San Francisco's Hackerspace, is having some hard times, so we're throwing an epic benefit and party this Saturday, to include eclectic performers, interactive art, a raffle and more! For more details, if any BBers want to put on demos or ideas share...",
			Enclosures: []models.Enclosure{},
		}},
	}

	assert := assert.New(t)

	select {
	case f := <-ch:
		assert.Equal(expected.FeedUrl, f.FeedUrl)
		assert.Equal(expected.WebsiteUrl, f.WebsiteUrl)
		assert.Equal(expected.FeedTitle, f.FeedTitle)
		assert.Equal(expected.FeedDescription, f.FeedDescription)
		assert.WithinDuration(expected.WhenLastUpdate.Time, f.WhenLastUpdate.Time, time.Second)
		assert.Equal(expected.Items, f.Items)

	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
