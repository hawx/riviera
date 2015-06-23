package feed

import (
	"errors"

	xmlx "github.com/jteeuwen/go-pkg-xmlx"
)

type Extension struct {
	Name      string
	Value     string
	Attrs     map[string]string
	Childrens map[string][]Extension
}

var days = map[string]int{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    7,
}

func readRss2(doc *xmlx.Document) (foundChannels []*Channel, err error) {
	const ns = "*"

	root := doc.SelectNode(ns, "rss")
	if root == nil {
		root = doc.SelectNode(ns, "RDF")
	}

	if root == nil {
		return foundChannels, errors.New("Failed to find rss/rdf node in XML.")
	}

	for _, node := range root.SelectNodes(ns, "channel") {
		ch := &Channel{
			Title:          node.S(ns, "title"),
			Description:    node.S(ns, "description"),
			Language:       node.S(ns, "language"),
			Copyright:      node.S(ns, "copyright"),
			ManagingEditor: node.S(ns, "managingEditor"),
			WebMaster:      node.S(ns, "webMaster"),
			PubDate:        node.S(ns, "pubDate"),
			LastBuildDate:  node.S(ns, "lastBuildDate"),
			Docs:           node.S(ns, "docs"),
			TTL:            node.I(ns, "ttl"),
			Rating:         node.S(ns, "rating"),
		}

		foundChannels = append(foundChannels, ch)

		for _, v := range node.SelectNodes(ns, "link") {
			lnk := Link{}
			if v.Name.Space == "http://www.w3.org/2005/Atom" && v.Name.Local == "link" {
				lnk.Href = v.As("", "href")
				lnk.Rel = v.As("", "rel")
				lnk.Type = v.As("", "type")
				lnk.HrefLang = v.As("", "hreflang")
			} else {
				lnk.Href = v.GetValue()
			}

			ch.Links = append(ch.Links, lnk)
		}

		for _, v := range node.SelectNodes(ns, "category") {
			ch.Categories = append(ch.Categories, &Category{
				Domain: v.As(ns, "domain"),
				Text:   v.GetValue(),
			})
		}

		if n := node.SelectNode(ns, "generator"); n != nil {
			ch.Generator = Generator{
				Text: n.GetValue(),
			}
		}

		for _, v := range node.SelectNodes(ns, "hour") {
			ch.SkipHours = append(ch.SkipHours, v.I(ns, "hour"))
		}

		for _, v := range node.SelectNodes(ns, "days") {
			ch.SkipDays = append(ch.SkipDays, days[v.GetValue()])
		}

		if n := node.SelectNode(ns, "image"); n != nil {
			ch.Image = Image{
				Title:       n.S(ns, "title"),
				Url:         n.S(ns, "url"),
				Link:        n.S(ns, "link"),
				Width:       n.I(ns, "width"),
				Height:      n.I(ns, "height"),
				Description: n.S(ns, "description"),
			}
		}

		if n := node.SelectNode(ns, "cloud"); n != nil {
			ch.Cloud = Cloud{
				Domain:            n.As(ns, "domain"),
				Port:              n.Ai(ns, "port"),
				Path:              n.As(ns, "path"),
				RegisterProcedure: n.As(ns, "registerProcedure"),
				Protocol:          n.As(ns, "protocol"),
			}
		}

		if n := node.SelectNode(ns, "textInput"); n != nil {
			ch.TextInput = Input{
				Title:       n.S(ns, "title"),
				Description: n.S(ns, "description"),
				Name:        n.S(ns, "name"),
				Link:        n.S(ns, "link"),
			}
		}

		list := node.SelectNodes(ns, "item")
		if len(list) == 0 {
			list = doc.SelectNodes(ns, "item")
		}

		for _, item := range list {
			i := &Item{
				Title:       item.S(ns, "title"),
				Description: item.S(ns, "description"),
				Comments:    item.S(ns, "comments"),
				PubDate:     item.S(ns, "pubDate"),
			}

			for _, v := range item.SelectNodes(ns, "link") {
				lnk := new(Link)
				if v.Name.Space == "http://www.w3.org/2005/Atom" && v.Name.Local == "link" {
					lnk.Href = v.As("", "href")
					lnk.Rel = v.As("", "rel")
					lnk.Type = v.As("", "type")
					lnk.HrefLang = v.As("", "hreflang")
				} else {
					lnk.Href = v.GetValue()
				}

				i.Links = append(i.Links, lnk)
			}

			if n := item.SelectNode(ns, "author"); n != nil {
				i.Author.Name = n.GetValue()

			} else if n := item.SelectNode(ns, "creator"); n != nil {
				i.Author.Name = n.GetValue()
			}

			if n := item.SelectNode(ns, "guid"); n != nil {
				i.Guid = &Guid{Guid: n.GetValue(), IsPermaLink: n.As("", "isPermalink") == "true"}
			}

			for _, lv := range item.SelectNodes(ns, "category") {
				i.Categories = append(i.Categories, &Category{
					Domain: lv.As(ns, "domain"),
					Text:   lv.GetValue(),
				})
			}

			for _, lv := range item.SelectNodes(ns, "enclosure") {
				i.Enclosures = append(i.Enclosures, &Enclosure{
					Url:    lv.As(ns, "url"),
					Length: lv.Ai64(ns, "length"),
					Type:   lv.As(ns, "type"),
				})
			}

			if src := item.SelectNode(ns, "source"); src != nil {
				i.Source = new(Source)
				i.Source.Url = src.As(ns, "url")
				i.Source.Text = src.GetValue()
			}

			for _, lv := range item.SelectNodes("http://purl.org/rss/1.0/modules/content/", "*") {
				if lv.Name.Local == "encoded" {
					i.Content = &Content{
						Text: lv.String(),
					}
					break
				}
			}

			i.Extensions = make(map[string]map[string][]Extension)
			for _, lv := range item.Children {
				getExtensions(&i.Extensions, lv)
			}

			ch.Items = append(ch.Items, i)
		}

		ch.Extensions = make(map[string]map[string][]Extension)
		for _, v := range node.Children {
			getExtensions(&ch.Extensions, v)
		}

	}
	return foundChannels, err
}

func getExtensions(extensionsX *map[string]map[string][]Extension, node *xmlx.Node) {
	extentions := *extensionsX

	if ext, ok := getExtension(node); ok {
		if len(extentions[node.Name.Space]) == 0 {
			extentions[node.Name.Space] = make(map[string][]Extension, 0)
		}
		if len(extentions[node.Name.Space][node.Name.Local]) == 0 {
			extentions[node.Name.Space][node.Name.Local] = make([]Extension, 0)
		}
		extentions[node.Name.Space][node.Name.Local] = append(extentions[node.Name.Space][node.Name.Local], ext)
	}
}

func getExtension(node *xmlx.Node) (extension Extension, ok bool) {
	if node.Name.Space == "" {
		return extension, false
	}

	extension = Extension{
		Name:      node.Name.Local,
		Value:     node.GetValue(),
		Attrs:     make(map[string]string),
		Childrens: make(map[string][]Extension, 0),
	}

	for _, attr := range node.Attributes {
		extension.Attrs[attr.Name.Local] = attr.Value
	}

	for _, child := range node.Children {
		if ext, ok := getExtension(child); ok {
			extension.Childrens[child.Name.Local] = append(extension.Childrens[child.Name.Local], ext)
		}
	}

	return extension, true
}
