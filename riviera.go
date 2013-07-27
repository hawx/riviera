package main

import (
	"github.com/hoisie/web"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/kennygrant/sanitize"

	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

const DOCS = "http://scripting.com/stories/2010/12/06/innovationRiverOfNewsInJso.html"

type Wrapper struct {
	UpdatedFeeds Feeds    `json:"updatedFeeds"`
	Metadata     Metadata `json:"metadata"`
}

type Metadata struct {
	Docs      string `json:"docs"`
	WhenGMT   string `json:"whenGMT"`
	WhenLocal string `json:"whenLocal"`
	Version   string `json:"version"`
	Secs      int    `json:"secs,string"`
}

type Feeds struct {
	UpdatedFeeds []Feed `json:"updatedFeed"`
}

type Feed struct {
	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
	WhenLastUpdate  string `json:"whenLastUpdate"`
	Items           []Item `json:"item"`
}

type Item struct {
	Body      string `json:"body"`
	PermaLink string `json:"permaLink"`
	PubDate   string `json:"pubDate"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Id        int    `json:"id,string"`
}

func FetchList(urls ...string) Feeds {
	var wg sync.WaitGroup
	updatedFeeds := []Feed{}

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			updatedFeeds = append(updatedFeeds, Fetch(url)...)
		}(url)
	}

	wg.Wait()

	return Feeds{
		UpdatedFeeds: updatedFeeds,
	}
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
		if err != nil && parsed.Year() > 1 {
			return &parsed, nil
		}
	}

	return nil, errors.New("Time could not be parsed: " + dateStr)
}

func old(pubDate string) bool {
	date, err := parseTime(pubDate)
	if err != nil {
		log.Print(err)
		return true
	}

	lastWeek := time.Now().Add(-7 * 24 * time.Hour)
	return date.Before(lastWeek)
}

func Fetch(url string) []Feed {
	updatedFeeds := []Feed{}

	feed := new(rss.Feed)
	feed.CacheTimeout = 15
	feed.Type = "none"

	err := feed.Fetch(url, nil)
	if err != nil {
		log.Fatalf("[e] %s: %s\n", url, err)
	}

	id := 0
	for _, channel := range feed.Channels {
		timeNow := time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST")

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
			if old(item.PubDate) {
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

func asset(path, contentType string, ctx *web.Context) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	ctx.ContentType(contentType)
	return string(content)
}

func style(ctx *web.Context, path string) string {
	return asset("css/"+path, "css", ctx)
}

func script(ctx *web.Context, path string) string {
	return asset("js/"+path, "js", ctx)
}

func index(ctx *web.Context) string {
	return asset("index.html", "html", ctx)
}

func fetchRiver(ctx *web.Context) string {
	ctx.ContentType("js")

	now := time.Now()

	feeds := FetchList()

	elapsed := int(time.Since(now) / time.Second)
	timeGMT := time.Now().UTC().Format("Mon, 2 Jan 2006 15:04:05 MST")
	timeNow := time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST")

	metadata := Metadata{
		Docs:      DOCS,
		WhenGMT:   timeGMT,
		WhenLocal: timeNow,
		Version:   "3",
		Secs:      elapsed,
	}

	wrapper := Wrapper{
		Metadata:     metadata,
		UpdatedFeeds: feeds,
	}

	b, _ := json.Marshal(wrapper)

	return `onGetRiverStream(` + string(b) + `)`
}

func main() {
	web.Get("/css/(.*.css)", style)
	web.Get("/js/(.*.js)", script)
	web.Get("/river.js", fetchRiver)
	web.Get("/?", index)
	web.Run("0.0.0.0:9999")
}
