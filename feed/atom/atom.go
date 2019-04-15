// Package atom provides a parser for Atom feeds.
//
// Commentary taken from https://tools.ietf.org/html/rfc4287
package atom

import (
	"encoding/xml"
	"io"

	"hawx.me/code/riviera/feed/common"
)

// Parser is capable of reading Atom feeds.
type Parser struct{}

// CanRead returns true if the reader provides data that is XML and contains the
// expected namespace for an Atom feed.
func (Parser) CanRead(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) bool {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset

	var token xml.Token
	var err error
	for {
		if token, err = decoder.Token(); err != nil || token == nil {
			break
		}
		if t, ok := token.(xml.StartElement); ok {
			if t.Name.Space == "http://www.w3.org/2005/Atom" && t.Name.Local == "feed" {
				return true
			}
			break
		}
	}

	return false
}

func (Parser) Read(r io.Reader, charset func(charset string, input io.Reader) (io.Reader, error)) (foundChannels []*common.Channel, err error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset

	var feed atomFeed
	if err = decoder.Decode(&feed); err != nil {
		return
	}

	ch := &common.Channel{
		Title:         feed.Title.Text,
		LastBuildDate: feed.Updated,
		ID:            feed.ID,
		Rights:        feed.Rights,
	}

	for _, link := range feed.Links {
		ch.Links = append(ch.Links, common.Link{
			Href:     link.Href,
			Rel:      link.Rel,
			Type:     link.Type,
			HrefLang: link.HrefLang,
		})
	}

	if feed.SubTitle != nil {
		ch.SubTitle = common.SubTitle{
			Type: feed.SubTitle.Type,
			Text: feed.SubTitle.Text,
		}
	}

	if feed.Generator != nil {
		ch.Generator = common.Generator{
			URI:     feed.Generator.URI,
			Version: feed.Generator.Version,
			Text:    feed.Generator.Text,
		}
	}

	if len(feed.Authors) > 0 {
		ch.Author = common.Author{
			Name:  feed.Authors[0].Name,
			URI:   feed.Authors[0].URI,
			Email: feed.Authors[0].Email,
		}
	}

	for _, entry := range feed.Entries {
		i := &common.Item{
			Title:       entry.Title,
			ID:          entry.ID,
			PubDate:     entry.Updated,
			Description: entry.Summary,
		}

		for _, link := range entry.Links {
			if link.Rel == "enclosure" {
				i.Enclosures = append(i.Enclosures, common.Enclosure{
					URL:  link.Href,
					Type: link.Type,
				})
			} else {
				i.Links = append(i.Links, common.Link{
					Href:     link.Href,
					Rel:      link.Rel,
					Type:     link.Type,
					HrefLang: link.HrefLang,
				})
			}
		}

		for _, contributor := range entry.Contributors {
			i.Contributors = append(i.Contributors, contributor.Name)
		}

		for _, category := range entry.Categories {
			i.Categories = append(i.Categories, common.Category{
				Domain: "",
				Text:   category.Term,
			})
		}

		if entry.Content != nil {
			i.Content = &common.Content{
				Type: entry.Content.Type,
				Lang: entry.Content.Lang,
				Base: entry.Content.Base,
				Text: entry.Content.Text,
			}
		}

		if len(entry.Authors) > 0 {
			i.Author = common.Author{
				Name:  entry.Authors[0].Name,
				URI:   entry.Authors[0].URI,
				Email: entry.Authors[0].Email,
			}
		}

		if entry.MediaThumbnail != nil {
			i.Thumbnail = &common.Image{
				URL: entry.MediaThumbnail.URL,
			}

			if entry.MediaThumbnail.Width != nil {
				i.Thumbnail.Width = *entry.MediaThumbnail.Width
			}
			if entry.MediaThumbnail.Height != nil {
				i.Thumbnail.Height = *entry.MediaThumbnail.Height
			}
		}

		ch.Items = append(ch.Items, i)
	}

	foundChannels = append(foundChannels, ch)
	return
}

// The "atom:feed" element is the document (i.e., top-level) element of an Atom
// Feed Document, acting as a container for metadata and data associated with
// the feed.  Its element children consist of metadata elements followed by zero
// or more atom:entry child elements.
type atomFeed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`

	// atom:feed elements MUST contain one or more atom:author elements, unless
	// all of the atom:feed element's child atom:entry elements contain at least
	// one atom:author element.
	Authors []atomAuthor `xml:"http://www.w3.org/2005/Atom author"`

	// atom:feed elements MAY contain any number of atom:category elements.
	Categories []atomCategory `xml:"http://www.w3.org/2005/Atom category"`

	// atom:feed elements MAY contain any number of atom:contributor elements.
	Contributors []atomContributor `xml:"http://www.w3.org/2005/Atom contributor"`

	// atom:feed elements MUST NOT contain more than one atom:generator element.
	Generator *atomGenerator `xml:"http://www.w3.org/2005/Atom generator"`

	// atom:feed elements MUST NOT contain more than one atom:icon element.
	// don't care

	// atom:feed elements MUST NOT contain more than one atom:logo element.
	// don't care

	// atom:feed elements MUST contain exactly one atom:id element.
	ID string `xml:"http://www.w3.org/2005/Atom id"`

	// atom:feed elements SHOULD contain one atom:link element with a rel
	// attribute value of "self".  This is the preferred URI for retrieving Atom
	// Feed Documents representing this Atom feed.
	//
	// atom:feed elements MUST NOT contain more than one atom:link element with a
	// rel attribute value of "alternate" that has the same combination of type
	// and hreflang attribute values.
	//
	// atom:feed elements MAY contain additional atom:link elements beyond those
	// described above.
	Links []atomLink `xml:"http://www.w3.org/2005/Atom link"`

	// atom:feed elements MUST NOT contain more than one atom:rights element.
	Rights string `xml:"http://www.w3.org/2005/Atom rights"`

	// atom:feed elements MUST NOT contain more than one atom:subtitle element.
	SubTitle *atomSubTitle `xml:"http://www.w3.org/2005/Atom subtitle"`

	// atom:feed elements MUST contain exactly one atom:title element.
	Title atomTitle `xml:"http://www.w3.org/2005/Atom title"`

	// atom:feed elements MUST contain exactly one atom:updated element.
	Updated string `xml:"http://www.w3.org/2005/Atom updated"`

	Entries []atomEntry `xml:"http://www.w3.org/2005/Atom entry"`
}

type atomTitle struct {
	Type string `xml:"http://www.w3.org/2005/Atom type,attr"`
	Text string `xml:",chardata"`
}

type atomLink struct {
	Href     string `xml:"href,attr"`
	Rel      string `xml:"rel,attr"`
	Type     string `xml:"type,attr"`
	HrefLang string `xml:"hreflang,attr"`
}

type atomSubTitle struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

type atomGenerator struct {
	URI     string `xml:"uri,attr"`
	Version string `xml:"version,attr"`
	Text    string `xml:",chardata"`
}

type atomAuthor struct {
	Name  string `xml:"http://www.w3.org/2005/Atom name"`
	URI   string `xml:"http://www.w3.org/2005/Atom uri"`
	Email string `xml:"http://www.w3.org/2005/Atom email"`
}

// The "atom:entry" element represents an individual entry, acting as a
// container for metadata and data associated with the entry.  This element can
// appear as a child of the atom:feed element, or it can appear as the document
// (i.e., top-level) element of a stand-alone Atom Entry Document.
type atomEntry struct {
	// atom:entry elements MUST contain one or more atom:author elements, unless
	// the atom:entry contains an atom:source element that contains an atom:author
	// element or, in an Atom Feed Document, the atom:feed element contains an
	// atom:author element itself.
	Authors []atomAuthor `xml:"http://www.w3.org/2005/Atom author"`

	// atom:entry elements MAY contain any number of atom:category elements.
	Categories []atomCategory `xml:"http://www.w3.org/2005/Atom category"`

	// atom:entry elements MUST NOT contain more than one atom:content element.
	Content *atomContent `xml:"http://www.w3.org/2005/Atom content"`

	// atom:entry elements MAY contain any number of atom:contributor elements.
	Contributors []atomContributor `xml:"http://www.w3.org/2005/Atom contributor"`

	// atom:entry elements MUST contain exactly one atom:id element.
	ID string `xml:"http://www.w3.org/2005/Atom id"`

	// atom:entry elements that contain no child atom:content element MUST contain
	// at least one atom:link element with a rel attribute value of "alternate".
	//
	// atom:entry elements MUST NOT contain more than one atom:link element with a
	// rel attribute value of "alternate" that has the same combination of type
	// and hreflang attribute values.
	//
	// atom:entry elements MAY contain additional atom:link elements beyond those
	// described above.
	Links []atomLink `xml:"http://www.w3.org/2005/Atom link"`

	// atom:entry elements MUST NOT contain more than one atom:published element.
	// I don't care?

	// atom:entry elements MUST NOT contain more than one atom:rights element.
	// I don't care?

	// atom:entry elements MUST NOT contain more than one atom:source element.
	// I don't care?

	// atom:entry elements MUST contain an atom:summary element in either of the
	// following cases:
	//
	//   * the atom:entry contains an atom:content that has a "src" attribute (and
	//     is thus empty).
	//
	//   * the atom:entry contains content that is encoded in Base64; i.e., the
	//     "type" attribute of atom:content is a MIME media type [MIMEREG], but is
	//     not an XML media type [RFC3023], does not begin with "text/", and does
	//     not end with "/xml" or "+xml".
	//
	// atom:entry elements MUST NOT contain more than one atom:summary element.
	Summary string `xml:"http://www.w3.org/2005/Atom summary"`

	// atom:entry elements MUST contain exactly one atom:title element.
	Title string `xml:"http://www.w3.org/2005/Atom title"`

	// atom:entry elements MUST contain exactly one atom:updated element.
	Updated string `xml:"http://www.w3.org/2005/Atom updated"`

	// http://www.rssboard.org/media-rss#media-thumbnails
	//
	// Allows particular images to be used as representative images for the media
	// object. If multiple thumbnails are included, and time coding is not at
	// play, it is assumed that the images are in order of importance. It has one
	// required attribute and three optional attributes.
	MediaThumbnail *struct {
		// url specifies the url of the thumbnail. It is a required attribute.
		URL string `xml:"url,attr"`

		// height specifies the height of the thumbnail. It is an optional attribute.
		Height *int `xml:"height,attr"`

		// width specifies the width of the thumbnail. It is an optional attribute.
		Width *int `xml:"width,attr"`

		// time specifies the time offset in relation to the media object. Typically
		// this is used when creating multiple keyframes within a single video. The
		// format for this attribute should be in the DSM-CC's Normal Play Time
		// (NTP) as used in RTSP [RFC 2326 3.6 Normal Play Time]. It is an optional
		// attribute.
		Time *string `xml:"time,attr"`
	} `xml:"http://search.yahoo.com/mrss/ thumbnail"`
}

type atomContributor struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	// The "term" attribute is a string that identifies the category to which the
	// entry or feed belongs.  Category elements MUST have a "term" attribute.
	Term string `xml:"term,attr"`

	// The "scheme" attribute is an IRI that identifies a categorization scheme.
	// Category elements MAY have a "scheme" attribute.
	Scheme string `xml:"scheme,attr"`

	// The "label" attribute provides a human-readable label for display in
	// end-user applications.  Category elements MAY have a "label" attribute.
	Label string `xml:"label,attr"`
}

// The "atom:content" element either contains or links to the content of the
// entry.  The content of atom:content is Language-Sensitive.
type atomContent struct {
	Type string `xml:"type,attr"`
	Lang string `xml:"xml lang,attr"`
	Base string `xml:"xml base,attr"`
	Text string `xml:",chardata"`
}
