package river

type Wrapper struct {
	UpdatedFeeds Feeds    `json:"updatedFeeds"`
	Metadata     Metadata `json:"metadata"`
}

type Metadata struct {
	Docs      string  `json:"docs"`
	WhenGMT   string  `json:"whenGMT"`
	WhenLocal string  `json:"whenLocal"`
	Version   string  `json:"version"`
	Secs      float64 `json:"secs,string"`
}

type Feeds struct {
	UpdatedFeeds []Feed `json:"updatedFeed"`
}

type Feed struct {
	FeedUrl         string `json:"feedUrl"`
	WebsiteUrl      string `json:"websiteUrl"`
	FeedTitle       string `json:"feedTitle"`
	FeedDescription string `json:"feedDescription"`
	WhenLastUpdate  string `json:"whenLastUpdate"`
	Items           []Item `json:"item"`
}

type Item struct {
	Body      string `json:"body"`
	PermaLink string `json:"permaLink"`
	PubDate   string `json:"pubDate"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Id        int    `json:"id,string"`
}
