package common

type Channel struct {
	Categories     []Category
	Cloud          Cloud
	Copyright      string
	Description    string
	Docs           string
	Extensions     map[string]map[string][]Extension
	Generator      Generator
	Image          Image
	Items          []*Item
	Language       string
	LastBuildDate  string
	Links          []Link
	ManagingEditor string
	PubDate        string
	Rating         string
	SkipDays       []int
	SkipHours      []int
	TTL            int
	TextInput      Input
	Title          string
	WebMaster      string

	// Atom fields
	Author   Author
	ID       string
	Rights   string
	SubTitle SubTitle
}

func (c *Channel) Key() string {
	switch {
	case len(c.ID) != 0:
		return c.ID
	default:
		return c.Title
	}
}
