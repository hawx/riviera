package feed

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
)

func itemHandler(feed *Feed, ch *Channel, newitems []*Item) {}

func charsetReader(name string, r io.Reader) (io.Reader, error) {
	return charset.NewReader(name, r)
}

func TestFeed(t *testing.T) {
	feedlist := []string{
		"/testdata/cyber.law.harvard.edu-sampleRss091.xml", // "http://cyber.law.harvard.edu/rss/examples/sampleRss091.xml", // Non-utf8 encoding.
		"/testdata/store.steampowered.com-news.xml",        // "http://store.steampowered.com/feeds/news.xml", // This feed violates the rss spec.
		"/testdata/cyber.law.harvard.edu-sampleRss092.xml", // "http://cyber.law.harvard.edu/rss/examples/sampleRss092.xml",
		"/testdata/cyber.law.harvard.edu-rss2sample.xml",   // "http://cyber.law.harvard.edu/rss/examples/rss2sample.xml",
		"/testdata/blog.case.edu-feed.atom",                // "http://blog.case.edu/news/feed.atom",
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.Open(r.URL.Path[1:])
		io.Copy(w, f)
		f.Close()
	}))

	var feed *Feed
	var err error

	for _, uri := range feedlist {
		feed = New(5, itemHandler, NewDatabase())

		if _, err = feed.Fetch(s.URL+uri, http.DefaultClient, charset.NewReader); err != nil {
			t.Errorf("%s >>> %s", uri, err)
			return
		}
	}
}

func Test_NewItem(t *testing.T) {
	file, _ := os.Open("testdata/initial.atom")

	itemsCh := make(chan []*Item, 2)
	feed := New(1, func(_ *Feed, _ *Channel, newitems []*Item) {
		itemsCh <- newitems
	}, NewDatabase())
	err := feed.load(file, nil)
	if err != nil {
		t.Error(err)
	}

	file.Close()

	file, _ = os.Open("testdata/initial_plus_one_new.atom")
	defer file.Close()
	feed.load(file, nil)
	expected := "Second title"

	select {
	case items := <-itemsCh:
		if len(items) != 1 {
			t.Errorf("Expected %d new item, got %d", 1, len(items))
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
			t.Errorf("Expected %d new item, got %d", 1, len(items))
		}

		if expected != items[0].Title {
			t.Errorf("Expected %s, got %s", expected, items[0].Title)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_AtomAuthor(t *testing.T) {
	file, err := os.Open("testdata/idownload.atom")
	if err != nil {
		t.Errorf("unable to load file")
	}
	defer file.Close()

	itemCh := make(chan *Item, 1)
	feed := New(1, func(f *Feed, ch *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	err = feed.load(file, nil)

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
	file, _ := os.Open("testdata/boing.rss")
	defer file.Close()

	itemCh := make(chan *Item, 1)
	feed := New(1, func(f *Feed, ch *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	feed.load(file, nil)

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
	file, _ := os.Open("testdata/extension.rss")
	defer file.Close()

	itemCh := make(chan *Item, 1)
	feed := New(1, func(_ *Feed, _ *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())
	feed.load(file, nil)

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
			t.Errorf("Expected files size to be %d but found %d", fileSizeExpected, len(files))
		}

		file := files[0]
		fileExpected := "cbmg_10qa.htm"
		if file.Attrs["file"] != fileExpected {
			t.Errorf("Expected file to be %s but found %d", fileExpected, len(file.Attrs["file"]))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_ChannelExtensions(t *testing.T) {
	file, _ := os.Open("testdata/extension.rss")
	defer file.Close()

	channelCh := make(chan *Channel, 1)
	feed := New(1, func(_ *Feed, ch *Channel, _ []*Item) {
		channelCh <- ch
	}, NewDatabase())

	feed.load(file, nil)

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
	file, _ := os.Open("testdata/iosBoardGameGeek.rss")
	defer file.Close()

	itemCh := make(chan *Item, 1)
	feed := New(1, func(_ *Feed, _ *Channel, newitems []*Item) {
		itemCh <- newitems[0]
	}, NewDatabase())

	feed.load(file, nil)

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
	file, _ := os.Open("testdata/nytimes.rss")
	defer file.Close()

	type pair struct {
		Item    *Item
		Channel *Channel
	}
	itemCh := make(chan pair, 1)

	feed := New(1, func(_ *Feed, ch *Channel, newitems []*Item) {
		itemCh <- pair{newitems[0], ch}
	}, NewDatabase())
	feed.load(file, nil)

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
