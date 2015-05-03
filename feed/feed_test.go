package feed

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
)

func charsetReader(name string, r io.Reader) (io.Reader, error) {
	return charset.NewReader(name, r)
}

func TestFeed(t *testing.T) {
	urilist := []string{
		"http://cyber.law.harvard.edu/rss/examples/sampleRss091.xml", // Non-utf8 encoding.
		"http://store.steampowered.com/feeds/news.xml",               // This feed violates the rss spec.
		"http://cyber.law.harvard.edu/rss/examples/sampleRss092.xml",
		"http://cyber.law.harvard.edu/rss/examples/rss2sample.xml",
		"http://blog.case.edu/news/feed.atom",
	}

	var feed *Feed
	var err error

	for _, uri := range urilist {
		feed = New(5, chanHandler, itemHandler, NewDatabase())

		if _, err = feed.Fetch(uri, &http.Client{Timeout: 5 * time.Second}, charsetReader); err != nil {
			t.Errorf("%s >>> %s", uri, err)
			return
		}
	}
}

func Test_NewItem(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/initial.atom")
	itemsCh := make(chan []*Item, 2)
	feed := New(1, chanHandler, func(_ *Feed, _ *Channel, newitems []*Item) {
		itemsCh <- newitems
	}, NewDatabase())
	err := feed.fetchBytes("http://example.com", content, nil)
	if err != nil {
		t.Error(err)
	}

	content, _ = ioutil.ReadFile("testdata/initial_plus_one_new.atom")
	feed.fetchBytes("http://example.com", content, nil)
	expected := "Second title"

	select {
	case items := <-itemsCh:
		if len(items) != 1 {
			t.Errorf("Expected %s new item, got %s", 1, len(items))
		}

		if "First title" != items[0].Title {
			t.Errorf("Expected %s, got %s", expected, items[0].Title)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	select {
	case items := <-itemsCh:
		if len(items) != 1 {
			t.Errorf("Expected %s new item, got %s", 1, len(items))
		}

		if expected != items[0].Title {
			t.Errorf("Expected %s, got %s", expected, items[0].Title)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_AtomAuthor(t *testing.T) {
	content, err := ioutil.ReadFile("testdata/idownload.atom")
	if err != nil {
		t.Errorf("unable to load file")
	}
	itemCh := make(chan *Item, 1)
	feed := New(1, chanHandler, func(f *Feed, ch *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	err = feed.fetchBytes("http://example.com", content, nil)

	select {
	case item := <-itemCh:
		expected := "Cody Lee"
		if item.Author.Name != expected {
			t.Errorf("Expected author to be %s but found %s", expected, item.Author.Name)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_RssAuthor(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/boing.rss")
	itemCh := make(chan *Item, 1)
	feed := New(1, chanHandler, func(f *Feed, ch *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	feed.fetchBytes("http://example.com", content, nil)

	select {
	case item := <-itemCh:
		expected := "Cory Doctorow"
		if item.Author.Name != expected {
			t.Errorf("Expected author to be %s but found %s", expected, item.Author.Name)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_ItemExtensions(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/extension.rss")
	itemCh := make(chan *Item, 1)
	feed := New(1, chanHandler, func(_ *Feed, _ *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	feed.fetchBytes("http://example.com", content, nil)

	select {
	case item := <-itemCh:
		edgarExtensionxbrlFiling := item.Extensions["http://www.sec.gov/Archives/edgar"]["xbrlFiling"][0].Childrens
		companyExpected := "Cellular Biomedicine Group, Inc."
		companyName := edgarExtensionxbrlFiling["companyName"][0]
		if companyName.Value != companyExpected {
			t.Errorf("Expected company to be %s but found %s", companyExpected, companyName.Value)
		}

		files := edgarExtensionxbrlFiling["xbrlFiles"][0].Childrens["xbrlFile"]
		fileSizeExpected := 10
		if len(files) != 10 {
			t.Errorf("Expected files size to be %s but found %s", fileSizeExpected, len(files))
		}

		file := files[0]
		fileExpected := "cbmg_10qa.htm"
		if file.Attrs["file"] != fileExpected {
			t.Errorf("Expected file to be %s but found %s", fileExpected, len(file.Attrs["file"]))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_ChannelExtensions(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/extension.rss")
	channelCh := make(chan *Channel, 1)
	feed := New(1, func(_ *Feed, newchannels []*Channel) {
		channelCh <- newchannels[0]
	}, itemHandler, NewDatabase())
	feed.fetchBytes("http://example.com", content, nil)

	select {
	case channel := <-channelCh:
		itunesExtentions := channel.Extensions["http://www.itunes.com/dtds/podcast-1.0.dtd"]

		authorExptected := "The Author"
		ownerEmailExpected := "test@rss.com"
		categoryExpected := "Politics"
		imageExptected := "http://golang.org/doc/gopher/project.png"

		if itunesExtentions["author"][0].Value != authorExptected {
			t.Errorf("Expected author to be %s but found %s", authorExptected, itunesExtentions["author"][0].Value)
		}

		if itunesExtentions["owner"][0].Childrens["email"][0].Value != ownerEmailExpected {
			t.Errorf("Expected owner email to be %s but found %s", ownerEmailExpected, itunesExtentions["owner"][0].Childrens["email"][0].Value)
		}

		if itunesExtentions["category"][0].Attrs["text"] != categoryExpected {
			t.Errorf("Expected category text to be %s but found %s", categoryExpected, itunesExtentions["category"][0].Attrs["text"])
		}

		if itunesExtentions["image"][0].Attrs["href"] != imageExptected {
			t.Errorf("Expected image href to be %s but found %s", imageExptected, itunesExtentions["image"][0].Attrs["href"])
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_CData(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/iosBoardGameGeek.rss")
	itemCh := make(chan *Item, 1)
	feed := New(1, chanHandler, func(_ *Feed, _ *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	feed.fetchBytes("http://example.com", content, nil)

	select {
	case item := <-itemCh:
		expected := `<p>abc<div>"def"</div>ghi`
		if item.Description != expected {
			t.Errorf("Expected item.Description to be [%s] but item.Description=[%s]", expected, item.Description)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_Link(t *testing.T) {
	content, _ := ioutil.ReadFile("testdata/nytimes.rss")
	type pair struct {
		Item    *Item
		Channel *Channel
	}
	itemCh := make(chan pair, 1)

	feed := New(1, chanHandler, func(_ *Feed, ch *Channel, newitems []*Item) {
		itemCh <- pair{newitems[0], ch}
	}, NewDatabase())
	feed.fetchBytes("http://example.com", content, nil)

	select {
	case p := <-itemCh:
		channel := p.Channel
		item := p.Item

		channelLinkExpected := "http://www.nytimes.com/services/xml/rss/nyt/US.xml"
		itemLinkExpected := "http://www.nytimes.com/2014/01/18/technology/in-keeping-grip-on-data-pipeline-obama-does-little-to-reassure-industry.html?partner=rss&emc=rss"

		if channel.Links[0].Href != channelLinkExpected {
			t.Errorf("Expected author to be %s but found %s", channelLinkExpected, channel.Links[0].Href)
		}

		if item.Links[0].Href != itemLinkExpected {
			t.Errorf("Expected author to be %s but found %s", itemLinkExpected, item.Links[0].Href)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func chanHandler(feed *Feed, newchannels []*Channel)        {}
func itemHandler(feed *Feed, ch *Channel, newitems []*Item) {}
