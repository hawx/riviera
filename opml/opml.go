// Package opml implements a functions capable of parsing opml files containing
// a list of feed subscriptions.
package opml

import (
	"encoding/xml"
	"io/ioutil"
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
	// Text    string `xml:"text,attr"`
	// Title   string `xml:"title,attr"`
	// Type    string `xml:"type,attr"`
	XmlUrl string `xml:"xmlUrl,attr"`
	// HtmlUrl string `xml:"htmlUrl,attr"`
}

func Parse(data []byte) (*Opml, error) {
	opml := Opml{}

	err := xml.Unmarshal(data, &opml)
	if err != nil {
		return nil, err
	}

	return &opml, nil
}

func Load(path string) (*Opml, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

func (doc *Opml) Save(path string) error {
	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	data = append([]byte(xml.Header), data...)
	return ioutil.WriteFile(path, data, 0644)
}
