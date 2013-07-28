package river

// Commentary copied from riverjs.org.

type Wrapper struct {
	UpdatedFeeds Feeds    `json:"updatedFeeds"`
	Metadata     Metadata `json:"metadata"`
}

type Feeds struct {
	UpdatedFeeds []Feed `json:"updatedFeed"`
}

// Each of these elements are mandatory, if there is no value for one it must be
// included with an empty string as its value.
//
// The elements all come from the top level of a feed, except for FeedUrl which
// is the address of the feed itself, and WhenLastUpdate which is the time when
// the new items from the feed were read by the aggregator.
type Feed struct {
	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
	WhenLastUpdate  string `json:"whenLastUpdate"`
	Items           []Item `json:"item"`
}

type Item struct {
	// Body is the description from the feed, with html markup stripped, and
	// limited to 280 characters. If the original text was more than the maximum
	// length, three periods are added to the end.
	Body      string `json:"body"`

	// Permalink, PubDate, Title and Link are straightforward copies of what
	// appeared in the feed.
	PermaLink string `json:"permaLink"`
	PubDate   string `json:"pubDate"`
	Title     string `json:"title"`
	Link      string `json:"link"`

	// Id is a number assigned to the item by the aggregator. Usuaully it is
	// increment by one for each item, but that's not guaranteed.
	Id        int    `json:"id,string"`

	// I am, at least for the moment, not including the optional elements that
	// river.js allows.
}

type Metadata struct {
	// Docs is a link to a web page that documents the format.
	Docs      string  `json:"docs"`

	// WhenGMT says when the file was built in a universal time.
	WhenGMT   string  `json:"whenGMT"`

	// WhenLocal says when the file was built in local time.
	WhenLocal string  `json:"whenLocal"`

	// Version is 3.
	Version   string  `json:"version"`

	// Secs is the number of seconds it took to build the file.
	Secs      float64 `json:"secs,string"`
}
