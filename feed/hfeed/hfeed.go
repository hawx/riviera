// Package hfeed provides a parser for h-feeds.
//
// See http://microformats.org/wiki/h-feed
package hfeed

import (
	htmlPkg "html"
	"io"
	"net/url"

	strip "github.com/grokify/html-strip-tags-go"
	"hawx.me/code/riviera/feed/common"
	"willnorris.com/go/microformats"
)

// Parser is capable of reading webpages marked up with the h-feed microformat.
type Parser struct{}

// CanRead returns true if the reader provides HTML containing the h-feed
// microformat.
func (Parser) CanRead(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) bool {
	data := microformats.Parse(r, nil)

	_, ok := findHFeed(data)
	return ok
}

func findHFeed(data *microformats.Data) (*microformats.Microformat, bool) {
	for _, item := range data.Items {
		if found, ok := findType(item, "h-feed"); ok {
			return found, ok
		}
	}

	return nil, false
}

func findType(item *microformats.Microformat, name string) (*microformats.Microformat, bool) {
	if contains(item.Type, "h-feed") {
		return item, true
	}

	for _, child := range item.Children {
		if found, ok := findType(child, name); ok {
			return found, ok
		}
	}

	return nil, false
}

func (p Parser) Read(r io.Reader, rootURL *url.URL, charset func(charset string, input io.Reader) (io.Reader, error)) (foundChannels []*common.Channel, err error) {
	data := microformats.Parse(r, rootURL)

	item, ok := findHFeed(data)
	if !ok {
		return
	}

	channel := &common.Channel{
		Links: []common.Link{
			{
				Href: rootURL.String(),
				Rel:  "alternate",
			},
			{
				Href: rootURL.String(),
				Rel:  "self",
			},
		},
	}

	if name, ok := getFirst(item.Properties, "name").(string); ok {
		channel.Title = name
	}

	for _, child := range item.Children {
		if contains(child.Type, "h-entry") {
			item := &common.Item{}

			if uid, ok := getFirst(child.Properties, "uid").(string); ok {
				item.GUID = &common.GUID{GUID: uid}
			}

			if name, ok := getFirst(child.Properties, "name").(string); ok {
				item.Title = name
			}

			if content, ok := getFirst(child.Properties, "content").(map[string]interface{}); ok {
				if text, ok := content["text"].(string); ok {
					if item.Title == text {
						item.Title = ""
					}

					item.Content = &common.Content{
						Text: text,
					}

				} else if html, ok := content["html"].(string); ok {
					html = htmlPkg.UnescapeString(strip.StripTags(html))
					if html == item.Title {
						item.Title = ""
					}

					item.Content = &common.Content{
						Text: html,
					}
				}
			}

			if url, ok := getFirst(child.Properties, "url").(string); ok {
				item.Links = []common.Link{{
					Href: url,
					Rel:  "alternate",
				}}
			}

			if published, ok := getFirst(child.Properties, "published").(string); ok {
				item.PubDate = published
			}

			channel.Items = append(channel.Items, item)
		}
	}

	foundChannels = append(foundChannels, channel)

	return
}

func contains(list []string, needle string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}

	return false
}

func getFirst(props map[string][]interface{}, name string) interface{} {
	list, ok := props[name]
	if !ok || len(list) == 0 {
		return nil
	}

	return list[0]
}
