// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/kennygrant/sanitize"
	"github.com/nu7hatch/gouuid"
	"encoding/json"
	"errors"
	"log"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type River interface {
	Latest() []Feed
	Close()
}

func New(uris []string, cutOff time.Duration) River {
	rivers := map[string]River{}

	for _, uri := range uris {
		rivers[uri] = newPoller(uri, cutOff)
	}

	return &collater{rivers}
}

func Build(river River) string {
	return fromFeeds(river.Latest())
}

func fromFeeds(feeds []Feed) string {
	updatedFeeds := Feeds{feeds}
	start := time.Now()

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
