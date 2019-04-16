// Package feed provides a feed fetcher capable of reading multiple formats into
// a common structure.
package feed

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"hawx.me/code/riviera/feed/atom"
	"hawx.me/code/riviera/feed/common"
	"hawx.me/code/riviera/feed/hfeed"
	"hawx.me/code/riviera/feed/jsonfeed"
	"hawx.me/code/riviera/feed/rdf"
	"hawx.me/code/riviera/feed/rss"
)

const userAgent = "riviera golang"

// ItemHandler is a callback function invoked when a feed has been fetched.
type ItemHandler func(f *Feed, ch *common.Channel, newitems []*common.Item)

// Feed manages polling of a web feed, either in atom, rss, rdf or jsonfeed format.
type Feed struct {
	// Custom cache timeout.
	cacheTimeout time.Duration

	// Type of feed. Rss, Atom, etc
	format string

	// Channels with content.
	channels []*common.Channel

	// URL from which this feed was created.
	uri *url.URL

	// Known containing a list of known Items and Channels for this instance
	known Database

	// A notification function, used to notify the host when a new item
	// has been found for a given channel.
	itemhandler ItemHandler

	// Last time content was fetched. Used in conjunction with CacheTimeout
	// to ensure we don't get content too often.
	lastupdate time.Time

	// The latest value of the ETag header returned from the last fetch.
	eTag string
}

// New creates a new feed that can be polled for updates.
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
// The charset parameter overrides the xml decoder's CharsetReader. This allows
// us to specify a custom character encoding conversion routine when dealing
// with non-utf8 input. Supply 'nil' to use the default from Go's xml package.
//
// The client parameter allows the use of arbitrary network connections, for
// example the Google App Engine "URL Fetch" service.
//
// If the feed is unable to update (see CanUpdate) then no request will be made,
// instead the result will be (status=-1, err=nil).
func (f *Feed) Fetch(uri string, client *http.Client, charset func(charset string, input io.Reader) (io.Reader, error)) (status int, err error) {
	if !f.CanUpdate() {
		return -1, nil
	}

	f.uri, _ = url.Parse(uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return -1, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("If-Modified-Since", f.lastupdate.Format(time.RFC1123))
	if f.eTag != "" {
		req.Header.Set("If-None-Match", f.eTag)
	}

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}

	f.eTag = resp.Header.Get("ETag")

	return resp.StatusCode, f.load(resp.Body, charset)
}

func (f *Feed) load(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) (err error) {
	f.channels, err = Parse(r, f.uri, charset)
	if err != nil || len(f.channels) == 0 {
		return
	}

	// reset cache timeout values according to feed specified values (TTL)
	if f.cacheTimeout < time.Minute*time.Duration(f.channels[0].TTL) {
		f.cacheTimeout = time.Minute * time.Duration(f.channels[0].TTL)
	}

	f.notifyListeners()
	return
}

// ErrUnsupportedFormat is returned when a feed is encountered in a format that
// is not understood.
var ErrUnsupportedFormat = errors.New("Unsupported feed")

var parsers = []common.Parser{
	atom.Parser{},
	rss.Parser{},
	rdf.Parser{},
	jsonfeed.Parser{},
	hfeed.Parser{},
}

// Parse reads the content from the provided reader, returning any feed channels
// found. If the feed is of a format not supported it will return
// ErrUnsupportedFormat.
func Parse(r io.Reader, rootURL *url.URL, charset func(charset string, input io.Reader) (io.Reader, error)) (chs []*common.Channel, err error) {
	data, _ := ioutil.ReadAll(r)
	br := bytes.NewReader(data)

	for _, parser := range parsers {
		if parser.CanRead(br, charset) {
			if _, err := br.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			return parser.Read(br, rootURL, charset)
		}
		if _, err := br.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	return nil, ErrUnsupportedFormat
}

func (f *Feed) notifyListeners() {
	for _, channel := range f.channels {
		var newitems []*common.Item

		for _, item := range channel.Items {
			if !f.known.Contains(item.Key()) {
				newitems = append(newitems, item)
			}
		}

		if len(newitems) > 0 && f.itemhandler != nil {
			f.itemhandler(f, channel, newitems)
		}
	}
}

// CanUpdate returns true or false depending on whether the CacheTimeout value
// has expired or not. Additionally, it will ensure that we adhere to the RSS
// spec's SkipDays and SkipHours values. If this function returns true, you can
// be sure that a fresh feed update will be performed.
func (f *Feed) CanUpdate() bool {
	// Make sure we are not within the specified cache-limit.
	// This ensures we don't request data too often.
	utc := time.Now().UTC()
	if utc.Sub(f.lastupdate) < f.cacheTimeout {
		return false
	}

	// If skipDays or skipHours are set in the RSS feed, use these to see if
	// we can update.
	if len(f.channels) == 1 && f.format == "rss" {
		if len(f.channels[0].SkipDays) > 0 {
			for _, v := range f.channels[0].SkipDays {
				if time.Weekday(v) == utc.Weekday() {
					return false
				}
			}
		}

		if len(f.channels[0].SkipHours) > 0 {
			for _, v := range f.channels[0].SkipHours {
				if v == utc.Hour() {
					return false
				}
			}
		}
	}

	f.lastupdate = utc
	return true
}

// DurationTillUpdate returns the number of seconds needed to elapse before the
// feed should update.
func (f *Feed) DurationTillUpdate() time.Duration {
	return f.cacheTimeout - time.Now().UTC().Sub(f.lastupdate)
}
