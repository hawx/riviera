// Package opml implements functions capable of parsing opml files containing a
// list of feed subscriptions.
package opml

import (
	"encoding/xml"
	"io"
	"os"
)

type Opml struct {
	XMLName struct{} `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outline []Outline `xml:"outline"`
}

type Outline struct {
	// Comments: http://dev.opml.org/spec2.html
	//
	// Required attributes: type, text, xmlUrl. For outline elements whose type is
	// rss, the text attribute should initially be the top-level title element in
	// the feed being pointed to, however since it is user-editable, processors
	// should not depend on it always containing the title of the feed. xmlUrl is
	// the http address of the feed.

	Type   string `xml:"type,attr,omitempty"`
	Text   string `xml:"text,attr,omitempty"`
	XmlUrl string `xml:"xmlUrl,attr,omitempty"`

	// Optional attributes: description, htmlUrl, language, title. These
	// attributes are useful when presenting a list of subscriptions to a user,
	// except for version, they are all derived from information in the feed
	// itself.

	// description is the top-level description element from the feed.
	Description string `xml:"description,attr,omitempty"`

	// htmlUrl is the top-level link element.
	HtmlUrl string `xml:"htmlUrl,attr,omitempty"`

	// language is the value of the top-level language element.
	Language string `xml:"language,attr,omitempty"`

	// title is probably the same as text, it should not be omitted. title
	// contains the top-level title element from the feed.
	Title string `xml:"title,attr,omitempty"`
}

func Load(path string) (doc Opml, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	err = xml.NewDecoder(file).Decode(&doc)
	if err != nil {
		return
	}

	return
}

func (doc Opml) WriteTo(w io.Writer) {
	xml.NewEncoder(w).Encode(doc)
}
