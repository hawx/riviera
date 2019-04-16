// Package hfeed provides a parser for h-feeds.
//
// See http://microformats.org/wiki/h-feed
package hfeed

import (
	"io"
	"log"

	"hawx.me/code/riviera/feed/common"
	"willnorris.com/go/microformats"
)

// Parser is capable of reading webpages marked up with the h-feed microformat.
type Parser struct{}

// CanRead returns true if the reader provides HTML containing the h-feed
// microformat.
func (Parser) CanRead(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) bool {
	data := microformats.Parse(r, nil)

	for _, item := range data.Items {
		if contains(item.Type, "h-feed") {
			return true
		}
	}

	return false
}

func (Parser) Read(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) (foundChannels []*common.Channel, err error) {
	data := microformats.Parse(r, nil)

	for _, item := range data.Items {
		if contains(item.Type, "h-feed") {
			channel := &common.Channel{}
			log.Println(item.Properties)

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
						if html, ok := content["html"].(string); ok {
							item.Content = &common.Content{
								Text: html,
							}
						} else if text, ok := content["text"].(string); ok {
							item.Content = &common.Content{
								Text: text,
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

					log.Println(child.Properties)

					channel.Items = append(channel.Items, item)
				}
			}

			foundChannels = append(foundChannels, channel)
		}
	}

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
