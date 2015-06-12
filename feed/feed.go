/*
 Package feed provides an RSS and Atom feed fetcher.

 They are parsed into an object tree which is a hybrid of both the RSS and Atom
 standards.

 Supported feeds are:
 	- RSS v0.91, 0.91 and 2.0
 	- Atom 1.0

 The package allows us to maintain cache timeout management. This prevents
 querying the servers for feed updates too often. Apart from setting a cache
 timeout manually, the package also optionally adheres to the TTL, SkipDays and
 SkipHours values specified in RSS feeds.

 Because the object structure is a hybrid between both RSS and Atom specs, not
 all fields will be filled when requesting either an RSS or Atom feed. As many
 shared fields as possible are used but some of them simply do not occur in
 either the RSS or Atom spec.
*/
package feed

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	xmlx "github.com/jteeuwen/go-pkg-xmlx"
)

type ItemHandler func(f *Feed, ch *Channel, newitems []*Item)

type Feed struct {
	// Custom cache timeout.
	cacheTimeout time.Duration

	// Type of feed. Rss, Atom, etc
	format string

	// Version of the feed. Major and Minor.
	version [2]int

	// Channels with content.
	channels []*Channel

	// Url from which this feed was created.
	url string

	// Known containing a list of known Items and Channels for this instance
	known Database

	// A notification function, used to notify the host when a new item
	// has been found for a given channel.
	itemhandler ItemHandler

	// Last time content was fetched. Used in conjunction with CacheTimeout
	// to ensure we don't get content too often.
	lastupdate time.Time
}

func New(cachetimeout time.Duration, ih ItemHandler, database Database) *Feed {
	v := new(Feed)
	v.cacheTimeout = cachetimeout
	v.format = "none"
	v.known = database
	v.itemhandler = ih
	return v
}

// Fetch retrieves the feed's latest content if necessary.
//
// The charset parameter overrides the xml decoder's CharsetReader.
// This allows us to specify a custom character encoding conversion
// routine when dealing with non-utf8 input. Supply 'nil' to use the
// default from Go's xml package.
//
// The client parameter allows the use of arbitrary network connections, for
// example the Google App Engine "URL Fetch" service.
func (this *Feed) Fetch(uri string, client *http.Client, charset xmlx.CharsetFunc) (int, error) {
	if !this.CanUpdate() {
		return -1, nil
	}

	this.url = uri

	r, err := client.Get(uri)
	if err != nil {
		return -1, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return r.StatusCode, nil
	}

	return r.StatusCode, this.load(r.Body, charset)
}

func (this *Feed) load(r io.Reader, charset xmlx.CharsetFunc) error {
	doc := xmlx.New()
	err := doc.LoadStream(r, charset)
	if err != nil {
		return err
	}

	return this.makeFeed(doc)
}

// fetchBytes retrieves the feed's content from the []byte
//
// The charset parameter overrides the xml decoder's CharsetReader.
// This allows us to specify a custom character encoding conversion
// routine when dealing with non-utf8 input. Supply 'nil' to use the
// default from Go's xml package.
func (this *Feed) fetchBytes(uri string, content []byte, charset xmlx.CharsetFunc) (err error) {
	this.url = uri

	doc := xmlx.New()

	if err = doc.LoadBytes(content, charset); err != nil {
		return
	}

	return this.makeFeed(doc)
}

func (this *Feed) makeFeed(doc *xmlx.Document) (err error) {
	// Extract type and version of the feed so we can have the appropriate
	// function parse it (rss 0.91, rss 0.92, rss 2, atom etc).
	this.format, this.version = this.GetVersionInfo(doc)

	if ok := this.testVersions(); !ok {
		err = errors.New(fmt.Sprintf("Unsupported feed: %s, version: %+v", this.format, this.version))
		return
	}

	if err = this.buildFeed(doc); err != nil || len(this.channels) == 0 {
		return
	}

	// reset cache timeout values according to feed specified values (TTL)
	if this.cacheTimeout < time.Minute*time.Duration(this.channels[0].TTL) {
		this.cacheTimeout = time.Minute * time.Duration(this.channels[0].TTL)
	}

	this.notifyListeners()

	return
}

func (this *Feed) notifyListeners() {
	for _, channel := range this.channels {
		var newitems []*Item

		for _, item := range channel.Items {
			if !this.known.Contains(item.Key()) {
				newitems = append(newitems, item)
			}
		}

		if len(newitems) > 0 && this.itemhandler != nil {
			this.itemhandler(this, channel, newitems)
		}
	}
}

// This function returns true or false, depending on whether the CacheTimeout
// value has expired or not. Additionally, it will ensure that we adhere to the
// RSS spec's SkipDays and SkipHours values. If this function returns true, you
// can be sure that a fresh feed update will be performed.
func (this *Feed) CanUpdate() bool {
	// Make sure we are not within the specified cache-limit.
	// This ensures we don't request data too often.
	utc := time.Now().UTC()
	if utc.Sub(this.lastupdate) < this.cacheTimeout {
		return false
	}

	// If skipDays or skipHours are set in the RSS feed, use these to see if
	// we can update.
	if len(this.channels) == 1 && this.format == "rss" {
		if len(this.channels[0].SkipDays) > 0 {
			for _, v := range this.channels[0].SkipDays {
				if time.Weekday(v) == utc.Weekday() {
					return false
				}
			}
		}

		if len(this.channels[0].SkipHours) > 0 {
			for _, v := range this.channels[0].SkipHours {
				if v == utc.Hour() {
					return false
				}
			}
		}
	}

	this.lastupdate = utc
	return true
}

// Returns the number of seconds needed to elapse
// before the feed should update.
func (this *Feed) DurationTillUpdate() time.Duration {
	return this.cacheTimeout - time.Now().UTC().Sub(this.lastupdate)
}

func (this *Feed) buildFeed(doc *xmlx.Document) (err error) {
	switch this.format {
	case "rss":
		err = this.readRss2(doc)
	case "atom":
		err = this.readAtom(doc)
	}
	return
}

func (this *Feed) testVersions() bool {
	switch this.format {
	case "rss":
		if this.version[0] > 2 || (this.version[0] == 2 && this.version[1] > 0) {
			return false
		}

	case "atom":
		if this.version[0] > 1 || (this.version[0] == 1 && this.version[1] > 0) {
			return false
		}

	default:
		return false
	}

	return true
}

func (this *Feed) GetVersionInfo(doc *xmlx.Document) (ftype string, fversion [2]int) {
	var node *xmlx.Node

	if node = doc.SelectNode("http://www.w3.org/2005/Atom", "feed"); node == nil {
		goto rss
	}

	ftype = "atom"
	fversion = [2]int{1, 0}
	return

rss:
	if node = doc.SelectNode("", "rss"); node != nil {
		ftype = "rss"
		version := node.As("", "version")
		p := strings.Index(version, ".")
		major, _ := strconv.Atoi(version[0:p])
		minor, _ := strconv.Atoi(version[p+1 : len(version)])
		fversion = [2]int{major, minor}
		return
	}

	// issue#5: Some documents have an RDF root node instead of rss.
	if node = doc.SelectNode("http://www.w3.org/1999/02/22-rdf-syntax-ns#", "RDF"); node != nil {
		ftype = "rss"
		fversion = [2]int{1, 1}
		return
	}

	ftype = "unknown"
	fversion = [2]int{0, 0}
	return
}
