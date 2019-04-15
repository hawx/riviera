// Package jsonfeed provides a parser for the jsonfeed format.
//
// See https://jsonfeed.org/version/1
package jsonfeed

import (
	"encoding/json"
	"io"

	"hawx.me/code/riviera/feed/common"
)

// Parser is capable of reading jsonfeed feeds.
type Parser struct{}

// CanRead returns true if the reader provides data that is JSON and contains
// the understood jsonfeed version of https://jsonfeed.org/version/1.
func (Parser) CanRead(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) bool {
	var feedVersion struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r).Decode(&feedVersion); err != nil {
		return false
	}

	return feedVersion.Version == "https://jsonfeed.org/version/1"
}

func (Parser) Read(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) (foundChannels []*common.Channel, err error) {
	var feed jsonFeed
	if err = json.NewDecoder(r).Decode(&feed); err != nil {
		return
	}

	ch := &common.Channel{
		Title:       feed.Title,
		Description: feed.Description,
	}

	if feed.Author != nil {
		ch.Author = common.Author{
			Name: feed.Author.Name,
			URI:  feed.Author.URL,
		}
	}

	if feed.HomePageURL != "" {
		ch.Links = append(ch.Links, common.Link{
			Href: feed.HomePageURL,
			Rel:  "alternate",
		})
	}
	if feed.FeedURL != "" {
		ch.Links = append(ch.Links, common.Link{
			Href: feed.FeedURL,
			Rel:  "self",
		})
	}

	for _, item := range feed.Items {
		i := &common.Item{
			Title:   item.Title,
			GUID:    &common.GUID{GUID: item.ID},
			PubDate: item.DatePublished,
		}

		if item.URL != "" {
			i.Links = append(i.Links, common.Link{
				Href: item.URL,
				Rel:  "alternate",
			})
		}
		if item.ExternalURL != "" {
			i.Links = append(i.Links, common.Link{
				Href: item.ExternalURL,
				Rel:  "related",
			})
		}

		if item.Author != nil {
			i.Author = common.Author{
				Name: item.Author.Name,
				URI:  item.Author.URL,
			}
		}

		if item.Summary != "" {
			i.Content = &common.Content{Text: item.Summary}
		} else if item.ContentText != "" {
			i.Content = &common.Content{Text: item.ContentText}
		} else if item.ContentHTML != "" {
			i.Content = &common.Content{Type: "html", Text: item.ContentHTML}
		}

		for _, attachment := range item.Attachments {
			i.Enclosures = append(i.Enclosures, common.Enclosure{
				URL:    attachment.URL,
				Length: attachment.SizeInBytes,
				Type:   attachment.MimeType,
			})
		}

		for _, tag := range item.Tags {
			i.Categories = append(i.Categories, common.Category{
				Text: tag,
			})
		}

		if item.Image != "" {
			i.Thumbnail = &common.Image{
				URL: item.Image,
			}
		}

		ch.Items = append(ch.Items, i)
	}

	foundChannels = append(foundChannels, ch)

	return
}

type jsonFeed struct {
	// version (required, string) is the URL of the version of the format the feed
	// uses. This should appear at the very top, though we recognize that not all
	// JSON generators allow for ordering.
	Version string `json:"version"`

	// title (required, string) is the name of the feed, which will often
	// correspond to the name of the website (blog, for instance), though not
	// necessarily.
	Title string `json:"title"`

	// home_page_url (optional but strongly recommended, string) is the URL of the
	// resource that the feed describes. This resource may or may not actually be
	// a “home” page, but it should be an HTML page. If a feed is published on the
	// public web, this should be considered as required. But it may not make
	// sense in the case of a file created on a desktop computer, when that file
	// is not shared or is shared only privately.
	HomePageURL string `json:"home_page_url"`

	// feed_url (optional but strongly recommended, string) is the URL of the
	// feed, and serves as the unique identifier for the feed. As with
	// home_page_url, this should be considered required for feeds on the public
	// web.
	FeedURL string `json:"feed_url"`

	// description (optional, string) provides more detail, beyond the title, on
	// what the feed is about. A feed reader may display this text.
	Description string `json:"description"`

	// icon (optional, string) is the URL of an image for the feed suitable to be
	// used in a timeline, much the way an avatar might be used. It should be
	// square and relatively large — such as 512 x 512 — so that it can be
	// scaled-down and so that it can look good on retina displays. It should use
	// transparency where appropriate, since it may be rendered on a non-white
	// background.
	Icon string `json:"icon"`

	// favicon (optional, string) is the URL of an image for the feed suitable to
	// be used in a source list. It should be square and relatively small, but not
	// smaller than 64 x 64 (so that it can look good on retina displays). As with
	// icon, this image should use transparency where appropriate, since it may be
	// rendered on a non-white background.
	Favicon string `json:"favicon"`

	// author (optional, object) specifies the feed author. The author object has
	// several members. These are all optional — but if you provide an author
	// object, then at least one is required:
	Author *jsonAuthor `json:"author"`

	// expired (optional, boolean) says whether or not the feed is finished — that
	// is, whether or not it will ever update again. A feed for a temporary event,
	// such as an instance of the Olympics, could expire. If the value is true,
	// then it’s expired. Any other value, or the absence of expired, means the
	// feed may continue to update.
	Expired *bool `json:"expired"`

	// items is an array, and is required.
	Items []jsonItem `json:"items"`
}

type jsonAuthor struct {
	// name (optional, string) is the author’s name.
	Name string `json:"name"`

	// url (optional, string) is the URL of a site owned by the author. It could
	// be a blog, micro-blog, Twitter account, and so on. Ideally the linked-to
	// page provides a way to contact the author, but that’s not required. The
	// URL could be a mailto: link, though we suspect that will be rare.
	URL string `json:"url"`

	// avatar (optional, string) is the URL for an image for the author. As with
	// icon, it should be square and relatively large — such as 512 x 512 — and
	// should use transparency where appropriate, since it may be rendered on a
	// non-white background.
	Avatar string `json:"avatar"`
}

type jsonItem struct {
	// id (required, string) is unique for that item for that feed over time. If
	// an item is ever updated, the id should be unchanged. New items should never
	// use a previously-used id. If an id is presented as a number or other type,
	// a JSON Feed reader must coerce it to a string. Ideally, the id is the full
	// URL of the resource described by the item, since URLs make great unique
	// identifiers.
	ID string `json:"id"`

	// url (optional, string) is the URL of the resource described by the
	// item. It’s the permalink. This may be the same as the id — but should be
	// present regardless.
	URL string `json:"url"`

	// external_url (very optional, string) is the URL of a page elsewhere. This
	// is especially useful for linkblogs. If url links to where you’re talking
	// about a thing, then external_url links to the thing you’re talking about.
	ExternalURL string `json:"external_url"`

	// title (optional, string) is plain text. Microblog items in particular may
	// omit titles.
	Title string `json:"title"`

	// content_html and content_text are each optional strings — but one or both
	// must be present. This is the HTML or plain text of the item. Important: the
	// only place HTML is allowed in this format is in content_html. A
	// Twitter-like service might use content_text, while a blog might use
	// content_html. Use whichever makes sense for your resource. (It doesn’t even
	// have to be the same for each item in a feed.)
	ContentHTML string `json:"content_html"`
	ContentText string `json:"content_text"`

	// summary (optional, string) is a plain text sentence or two describing the
	// item. This might be presented in a timeline, for instance, where a detail
	// view would display all of content_html or content_text.
	Summary string `json:"summary"`

	// image (optional, string) is the URL of the main image for the item. This
	// image may also appear in the content_html — if so, it’s a hint to the feed
	// reader that this is the main, featured image. Feed readers may use the
	// image as a preview (probably resized as a thumbnail and placed in a
	// timeline).
	Image string `json:"image"`

	// banner_image (optional, string) is the URL of an image to use as a
	// banner. Some blogging systems (such as Medium) display a different banner
	// image chosen to go with each post, but that image wouldn’t otherwise appear
	// in the content_html. A feed reader with a detail view may choose to show
	// this banner image at the top of the detail view, possibly with the title
	// overlaid.
	BannerImage string `json:"banner_image"`

	// date_published (optional, string) specifies the date in RFC 3339
	// format. (Example: 2010-02-07T14:04:00-05:00.)
	DatePublished string `json:"date_published"`

	// date_modified (optional, string) specifies the modification date in RFC
	// 3339 format.
	DateModified string `json:"date_modified"`

	// author (optional, object) has the same structure as the top-level
	// author. If not specified in an item, then the top-level author, if present,
	// is the author of the item.
	Author *jsonAuthor `json:"author"`

	// tags (optional, array of strings) can have any plain text values you
	// want. Tags tend to be just one word, but they may be anything. Note: they
	// are not the equivalent of Twitter hashtags. Some blogging systems and other
	// feed formats call these categories.
	Tags []string `json:"tags"`

	// attachments (optional, array) lists related resources. Podcasts, for
	// instance, would include an attachment that’s an audio or video file.
	Attachments []jsonAttachment `json:"attachments"`
}

type jsonAttachment struct {
	// url (required, string) specifies the location of the attachment.
	URL string `json:"url"`

	// mime_type (required, string) specifies the type of the attachment, such as
	// “audio/mpeg.”
	MimeType string `json:"mime_type"`

	// title (optional, string) is a name for the attachment. Important: if there
	// are multiple attachments, and two or more have the exact same title (when
	// title is present), then they are considered as alternate representations of
	// the same thing. In this way a podcaster, for instance, might provide an
	// audio recording in different formats.
	Title string `json:"title"`

	// size_in_bytes (optional, number) specifies how large the file is.
	SizeInBytes int64 `json:"size_in_bytes"`

	// duration_in_seconds (optional, number) specifies how long it takes to
	// listen to or watch, when played at normal speed.
	DurationInSeconds int `json:"duration_in_seconds"`
}
