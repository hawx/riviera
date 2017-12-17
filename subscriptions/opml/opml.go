// Package opml implements functions capable of parsing opml files containing a
// list of feed subscriptions.
package opml

import (
	"encoding/xml"
	"io"
	"os"

	"golang.org/x/net/html/charset"
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
	XMLURL string `xml:"xmlUrl,attr,omitempty"`

	// Optional attributes: description, htmlUrl, language, title. These
	// attributes are useful when presenting a list of subscriptions to a user,
	// except for version, they are all derived from information in the feed
	// itself.

	// description is the top-level description element from the feed.
	Description string `xml:"description,attr,omitempty"`

	// htmlUrl is the top-level link element.
	HTMLURL string `xml:"htmlUrl,attr,omitempty"`

	// language is the value of the top-level language element.
	Language string `xml:"language,attr,omitempty"`

	// title is probably the same as text, it should not be omitted. title
	// contains the top-level title element from the feed.
	Title string `xml:"title,attr,omitempty"`
}

// Load parses the OPML file at the path.
func Load(path string) (doc Opml, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	return Read(file)
}

// Read parses an OPML document.
func Read(r io.Reader) (doc Opml, err error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charset.NewReaderLabel
	err = d.Decode(&doc)
	return
}

// WriteTo writes the OPML document out.
func (doc Opml) WriteTo(w io.Writer) error {
	_, err := w.Write([]byte(xml.Header))
	if err != nil {
		return err
	}

	return xml.NewEncoder(w).Encode(doc)
}
