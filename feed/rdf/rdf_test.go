package rdf

import (
	"os"
	"testing"

	"golang.org/x/net/html/charset"

	"hawx.me/code/assert"
)

func TestRss09Feed(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/rss09.rdf")
	defer file.Close()

	channels, err := new(Parser).Read(file, charset.NewReaderLabel)
	assert.Nil(err)
	assert.Len(channels, 1)

	channel := channels[0]
	assert.Equal("Mozilla Dot Org", channel.Title)
	if assert.Len(channel.Links, 1) {
		assert.Equal("http://www.mozilla.org", channel.Links[0].Href)
	}
	assert.Equal("the Mozilla Organization\n    web site", channel.Description)

	assert.Equal("Mozilla", channel.Image.Title)
	assert.Equal("http://www.mozilla.org/images/moz.gif", channel.Image.Url)
	assert.Equal("http://www.mozilla.org", channel.Image.Link)

	if assert.Len(channel.Items, 5) {
		assert.Equal("New Status Updates", channel.Items[0].Title)
		assert.Equal("http://www.mozilla.org/status/", channel.Items[0].Links[0].Href)

		assert.Equal("Bugzilla Reorganized", channel.Items[1].Title)
		assert.Equal("http://www.mozilla.org/bugs/", channel.Items[1].Links[0].Href)

		assert.Equal("Mozilla Party, 2.0!", channel.Items[2].Title)
		assert.Equal("http://www.mozilla.org/party/1999/", channel.Items[2].Links[0].Href)

		assert.Equal("Unix Platform Parity", channel.Items[3].Title)
		assert.Equal("http://www.mozilla.org/build/unix.html", channel.Items[3].Links[0].Href)

		assert.Equal("NPL 1.0M published", channel.Items[4].Title)
		assert.Equal("http://www.mozilla.org/NPL/NPL-1.0M.html", channel.Items[4].Links[0].Href)
	}
}

func TestSteamFeed(t *testing.T) {
	assert := assert.New(t)

	file, _ := os.Open("testdata/steam-news.xml")
	defer file.Close()

	channels, err := new(Parser).Read(file, charset.NewReaderLabel)
	assert.Nil(err)
	assert.Len(channels, 1)

	channel := channels[0]
	assert.Equal("Steam RSS News Feed", channel.Title)
	if assert.Len(channel.Links, 1) {
		assert.Equal("http://www.steampowered.com/", channel.Links[0].Href)
	}
	assert.Equal("All Steam news, all the time!", channel.Description)

	if assert.Len(channel.Items, 20) {
		item := channel.Items[0]
		assert.Equal("Free Weekend - Fallout 4 - 67% Off", item.Title)
		if assert.Len(item.Links, 1) {
			assert.Equal("http://store.steampowered.com/news/29654/", item.Links[0].Href)
		}
		assert.Equal("2017-05-25T10:15:00-0700", item.PubDate)
		assert.Equal("Valve", item.Author.Name)
		if assert.Len(item.Categories, 1) {
			assert.Equal("Valve news update", item.Categories[0].Text)
		}
		assert.Equal(`Play <a href='http://store.steampowered.com/app/377160/'>Fallout 4</a> for FREE starting now through Sunday at 1PM Pacific Time. You can also pickup <a href='http://store.steampowered.com/app/377160/'>Fallout 4</a> at 67% off the regular price!*<br><br>If you already have Steam installed, <a href='steam://run/377160'>click here</a> to install or play Fallout 4.  If you don't have Steam, you can download it <a href='http://store.steampowered.com/about/'>here</a>.<br><br>*Offer ends Monday at 10AM Pacific Time<br><a href="http://store.steampowered.com/app/377160/"><img src="https://steamcdn-a.akamaihd.net/steam/apps/377160/capsule_467x181.jpg" style=" float: left; margin-right: 12px; height: 181px; width: 467px;"></a>`, item.Content.Text)

		item = channel.Items[1]
		assert.Equal("Weekend Deal - Don't Starve, 75% Off", item.Title)
		if assert.Len(item.Links, 1) {
			assert.Equal("http://store.steampowered.com/news/29643/", item.Links[0].Href)
		}
		assert.Equal("2017-05-25T10:05:00-0700", item.PubDate)
		assert.Equal("Valve", item.Author.Name)
		if assert.Len(item.Categories, 1) {
			assert.Equal("Valve news update", item.Categories[0].Text)
		}
		assert.Equal(`Save up to 75% on the <a href='http://store.steampowered.com/sale/dontstarve/'>Don't Starve series</a> as part of this week's Weekend Deal*!<br><br>*Offer ends Monday at 10AM Pacific Time<br><br><a href="http://store.steampowered.com/app/219740/"><img src="https://steamcdn-a.akamaihd.net/steam/apps/219740/capsule_467x181.jpg" style=" float: left; margin-right: 12px; height: 181px; width: 467px;"></a><br><br><a href="http://store.steampowered.com/app/322330/"><img src="https://steamcdn-a.akamaihd.net/steam/apps/322330/capsule_467x181.jpg" style=" float: left; margin-right: 12px; height: 181px; width: 467px;"></a>`, item.Content.Text)

		item = channel.Items[2]
		assert.Equal("Daily Deal - Glittermitten Grove, 40% Off", item.Title)
		if assert.Len(item.Links, 1) {
			assert.Equal("http://store.steampowered.com/news/29615/", item.Links[0].Href)
		}
		assert.Equal("2017-05-25T10:00:00-0700", item.PubDate)
		assert.Equal("Valve", item.Author.Name)
		if assert.Len(item.Categories, 1) {
			assert.Equal("Valve news update", item.Categories[0].Text)
		}
		assert.Equal(`Today's Deal: Save 40% on <a href='http://store.steampowered.com/app/536890/'>Glittermitten Grove</a>!*<br><br>Look for the deals each day on the front page of Steam.  Or follow us on <a href='http://twitter.com/steam_games'>twitter</a> or <a href='http://www.facebook.com/Steam'>Facebook</a> for instant notifications wherever you are!<br><br>*Offer ends Saturday at 10AM Pacific Time<br><a href="http://store.steampowered.com/app/536890/"><img src="https://steamcdn-a.akamaihd.net/steam/apps/536890/capsule_467x181.jpg" style=" float: left; margin-right: 12px; height: 181px; width: 467px;"></a>`, item.Content.Text)
	}
}
