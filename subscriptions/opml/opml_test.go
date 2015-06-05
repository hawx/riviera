package opml

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEmpty(t *testing.T) {
	r := strings.NewReader(``)

	doc, err := Read(r)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, Opml{}, doc)
}

func TestReadSubscriptions(t *testing.T) {
	r := strings.NewReader(`<?xml version="1.0" encoding="ISO-8859-1"?>
<opml version="2.0">
<head>
<title>mySubscriptions.opml</title>
<dateCreated>Sat, 18 Jun 2005 12:11:52 GMT</dateCreated>
<dateModified>Tue, 02 Aug 2005 21:42:48 GMT</dateModified>
<ownerName>Dave Winer</ownerName>
<ownerEmail>dave@scripting.com</ownerEmail>
<expansionState></expansionState>
<vertScrollState>1</vertScrollState>
<windowTop>61</windowTop>
<windowLeft>304</windowLeft>
<windowBottom>562</windowBottom>
<windowRight>842</windowRight>
</head>
<body>
<outline text="CNET News.com" description="Tech news and business reports by CNET News.com. Focused on information technology, core topics include computers, hardware, software, networking, and Internet media." htmlUrl="http://news.com.com/" language="unknown" title="CNET News.com" type="rss" version="RSS2" xmlUrl="http://news.com.com/2547-1_3-0-5.xml"/>
<outline text="washingtonpost.com - Politics" description="Politics" htmlUrl="http://www.washingtonpost.com/wp-dyn/politics?nav=rss_politics" language="unknown" title="washingtonpost.com - Politics" type="rss" version="RSS2" xmlUrl="http://www.washingtonpost.com/wp-srv/politics/rssheadlines.xml"/>
</body>
</opml>`)

	doc, err := Read(r)
	assert.Nil(t, err)
	assert.Equal(t, Opml{
		Version: "2.0",
		Head: Head{
			Title: "mySubscriptions.opml",
		},
		Body: Body{
			Outline: []Outline{
				{Type: "rss", Text: "CNET News.com", XmlUrl: "http://news.com.com/2547-1_3-0-5.xml", Description: "Tech news and business reports by CNET News.com. Focused on information technology, core topics include computers, hardware, software, networking, and Internet media.", HtmlUrl: "http://news.com.com/", Language: "unknown", Title: "CNET News.com"},
				{Type: "rss", Text: "washingtonpost.com - Politics", XmlUrl: "http://www.washingtonpost.com/wp-srv/politics/rssheadlines.xml", Description: "Politics", HtmlUrl: "http://www.washingtonpost.com/wp-dyn/politics?nav=rss_politics", Language: "unknown", Title: "washingtonpost.com - Politics"},
			},
		},
	}, doc)
}

func TestWrite(t *testing.T) {
	doc := Opml{
		Version: "2.0",
		Head: Head{
			Title: "mySubscriptions.opml",
		},
		Body: Body{
			Outline: []Outline{
				{Type: "rss", Text: "CNET News.com", XmlUrl: "http://news.com.com/2547-1_3-0-5.xml", Description: "Tech news and business reports by CNET News.com. Focused on information technology, core topics include computers, hardware, software, networking, and Internet media.", HtmlUrl: "http://news.com.com/", Language: "unknown", Title: "CNET News.com"},
				{Type: "rss", Text: "washingtonpost.com - Politics", XmlUrl: "http://www.washingtonpost.com/wp-srv/politics/rssheadlines.xml", Description: "Politics", HtmlUrl: "http://www.washingtonpost.com/wp-dyn/politics?nav=rss_politics", Language: "unknown", Title: "washingtonpost.com - Politics"},
			},
		},
	}

	var buf bytes.Buffer
	doc.WriteTo(&buf)

	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0"><head><title>mySubscriptions.opml</title></head><body><outline type="rss" text="CNET News.com" xmlUrl="http://news.com.com/2547-1_3-0-5.xml" description="Tech news and business reports by CNET News.com. Focused on information technology, core topics include computers, hardware, software, networking, and Internet media." htmlUrl="http://news.com.com/" language="unknown" title="CNET News.com"></outline><outline type="rss" text="washingtonpost.com - Politics" xmlUrl="http://www.washingtonpost.com/wp-srv/politics/rssheadlines.xml" description="Politics" htmlUrl="http://www.washingtonpost.com/wp-dyn/politics?nav=rss_politics" language="unknown" title="washingtonpost.com - Politics"></outline></body></opml>`, buf.String())
}
