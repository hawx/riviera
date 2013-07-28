// Package river implements a functions for generating river.js files. See
// riverjs.org for more information on the format.
package river

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/kennygrant/sanitize"

	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

func Build(callback string, cutOff time.Duration, urls ...string) string {
	start := time.Now()

	updatedFeeds := Fetch(cutOff, urls...)

	elapsed := time.Since(start).Seconds()
	now := time.Now()
	timeGMT := now.UTC().Format(time.RFC1123)
	timeNow := now.Format(time.RFC1123)

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

	return fmt.Sprintf("%s(%s)", callback, string(b))
}

func Fetch(cutOff time.Duration, urls ...string) Feeds {
	var wg sync.WaitGroup
	updatedFeeds := []Feed{}

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			updatedFeeds = append(updatedFeeds, fetchFromUrl(url, cutOff)...)
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

	id := 0
	for _, channel := range feed.Channels {
		timeNow := time.Now().Format(time.RFC1123)

		updatedFeed := Feed{
			FeedUrl:         url,
			FeedTitle:       channel.Title,
			FeedDescription: channel.Description,
			WhenLastUpdate:  timeNow,
			Items:           []Item{},
		}

		for _, link := range channel.Links {
			if updatedFeed.FeedUrl != "" && updatedFeed.WebsiteUrl != "" {
				break
			}

			if link.Rel == "self" {
				updatedFeed.FeedUrl = link.Href
			} else {
				updatedFeed.WebsiteUrl = link.Href
			}
		}

		for _, item := range channel.Items {
			if old(item.PubDate, cutOff) {
				break
			}

			i := Item{
				Body:      stripAndCrop(item.Description),
				PermaLink: item.Links[0].Href, // either this is wrong
				PubDate:   item.PubDate,
				Title:     item.Title,
				Link:      item.Links[0].Href, // or this is wrong
				Id:        id,
			}

			if feed.Type == "atom" {
				i.Body = stripAndCrop(item.Content.Text)
			}

			updatedFeed.Items = append(updatedFeed.Items, i)
			id++
		}
		updatedFeeds = append(updatedFeeds, updatedFeed)
	}

	return updatedFeeds
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

func old(pubDate string, cutOff time.Duration) bool {
	date, err := parseTime(pubDate)
	if err != nil {
		log.Print(err)
		return false
	}

	lastWeek := time.Now().Add(-cutOff)
	return date.Before(lastWeek)
}
