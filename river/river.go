// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/kennygrant/sanitize"
	uuid "github.com/nu7hatch/gouuid"

	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

func Build(cutOff time.Duration, urls ...string) string {
	start := time.Now()

	updatedFeeds := Fetch(cutOff, urls...)

	elapsed := time.Since(start).Seconds()
	now := time.Now()
	timeGMT := now.UTC().Format(time.RFC1123Z)
	timeNow := now.Format(time.RFC1123Z)

	metadata := Metadata{
		Docs:      DOCS,
		WhenGMT:   timeGMT,
		WhenLocal: timeNow,
		Version:   "3",
		Secs:      elapsed,
	}

	wrapper := Wrapper{
		Metadata:     metadata,
		UpdatedFeeds: updatedFeeds,
	}

	b, _ := json.Marshal(wrapper)

	return string(b)
}

func Fetch(cutOff time.Duration, urls ...string) Feeds {
	var wg sync.WaitGroup
	updatedFeeds := []Feed{}

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			updatedChannels := fetchFromUrl(url, cutOff)
			for _, channel := range updatedChannels {
				if len(channel.Items) > 0 {
					updatedFeeds = append(updatedFeeds, channel)
				}
			}
		}(url)
	}

	wg.Wait()

	return Feeds{
		UpdatedFeeds: updatedFeeds,
	}
}

func fetchFromUrl(url string, cutOff time.Duration) []Feed {
	updatedFeeds := []Feed{}

	feed := new(rss.Feed)
	feed.CacheTimeout = 15
	feed.Type = "none"

	err := feed.Fetch(url, nil)
	if err != nil {
		log.Fatalf("%s: %s\n", url, err)
	}

	for _, channel := range feed.Channels {
		updatedFeed := convertChannel(channel, url, cutOff)
		updatedFeeds = append(updatedFeeds, *updatedFeed)
	}

	return updatedFeeds
}

func Subscribe(url string, cutOff time.Duration) {

}

func convertChannel(channel *rss.Channel, url string, cutOff time.Duration) *Feed {
	f := &Feed{
  	FeedUrl: url,
	  FeedTitle: channel.Title,
	  FeedDescription: channel.Description,
	  WhenLastUpdate: time.Now().Format(time.RFC1123),
	  Items: []Item{},
	}

	for _, link := range channel.Links {
		if f.FeedUrl != "" && f.WebsiteUrl != "" { break }

		if link.Rel == "self" {
			f.FeedUrl = link.Href
		} else {
			f.WebsiteUrl = link.Href
		}
	}

	for _, item := range channel.Items {
		i := convertItem(item, cutOff)
		if i == nil { break }
		f.Items = append(f.Items, *i)
	}

	return f
}

func convertItem(item *rss.Item, cutOff time.Duration) *Item {
	pubDate, err := parseTime(item.PubDate)
	if err != nil {
		log.Fatal(err)
	}

	if old(pubDate, cutOff) { return nil }

	id, _ := uuid.NewV4()

	i := &Item{
    Body:      stripAndCrop(item.Description),
    PubDate:   pubDate.Format(time.RFC1123Z),
	  Title:     item.Title,
  	Id:        id.String(),
	}

	if len(item.Links) > 0 {
		i.PermaLink = item.Links[0].Href
		i.Link = item.Links[0].Href
	}

	if item.Content != nil {
		i.Body = stripAndCrop(item.Content.Text)
	}

	return i
}


// func PollFeed(uri string, timeout int) {
// 	feed := rss.New(timeout, true, chanHandler, itemHandler)

// 	for {
// 		if err := feed.Fetch(uri, nil); err != nil {
// 			fmt.Fprintf(os.Stderr, "[e] %s: %s", uri, err)
//                         return
// 		}

// 		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
// 	}
// }

// func channelHandler(feed *rss.Feed, newchannels []*rss.Channel) {
// 	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
// }

// func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
// 	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)
// }



// Strips html markup, then limits to 280 characters. If the original text was
// longer than 280 chars, three periods are appended.
func stripAndCrop(content string) string {
	content = sanitize.HTML(content)

	if len(content) < 280 {
		return content
	}

	return content[0:280] + "..."
}

func parseTime(dateStr string) (*time.Time, error) {
	formats := []string{
		time.RFC822,   // RSS style
		time.RFC822Z,  // RSS style, with numeric time zone
		time.RFC1123,  // RSS style, with day
		time.RFC1123Z, // RSS style, with day and numeric time zone
		time.RFC3339,  // ATOM style
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, dateStr)
		if err == nil && parsed.Year() > 1 {
			return &parsed, nil
		}
	}

	return nil, errors.New("Time could not be parsed: " + dateStr)
}

func old(pubDate *time.Time, cutOff time.Duration) bool {
	lastWeek := time.Now().Add(-cutOff)
	return pubDate.Before(lastWeek)
}
