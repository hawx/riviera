package feed

import xmlx "github.com/jteeuwen/go-pkg-xmlx"

func readAtom(doc *xmlx.Document) (foundChannels []*Channel, err error) {
	const ns = "http://www.w3.org/2005/Atom"

	for _, node := range doc.SelectNodes(ns, "feed") {
		foundChannels = append(foundChannels, readAtomChannel(ns, node))
	}

	return foundChannels, err
}

func readAtomChannel(ns string, node *xmlx.Node) *Channel {
	ch := &Channel{
		Title:         node.S(ns, "title"),
		LastBuildDate: node.S(ns, "updated"),
		Id:            node.S(ns, "id"),
		Rights:        node.S(ns, "rights"),
	}

	for _, v := range node.SelectNodesDirect(ns, "link") {
		ch.Links = append(ch.Links, Link{
			Href:     v.As("", "href"),
			Rel:      v.As("", "rel"),
			Type:     v.As("", "type"),
			HrefLang: v.As("", "hreflang"),
		})
	}

	if tn := node.SelectNode(ns, "subtitle"); tn != nil {
		ch.SubTitle = SubTitle{
			Type: tn.As("", "type"),
			Text: tn.GetValue(),
		}
	}

	if tn := node.SelectNode(ns, "generator"); tn != nil {
		ch.Generator = Generator{
			Uri:     tn.As("", "uri"),
			Version: tn.As("", "version"),
			Text:    tn.GetValue(),
		}
	}

	if tn := node.SelectNode(ns, "author"); tn != nil {
		ch.Author = Author{
			Name:  tn.S("", "name"),
			Uri:   tn.S("", "uri"),
			Email: tn.S("", "email"),
		}
	}

	for _, item := range node.SelectNodes(ns, "entry") {
		ch.Items = append(ch.Items, readAtomItem(ns, item))
	}

	return ch
}

func readAtomItem(ns string, item *xmlx.Node) *Item {
	i := &Item{
		Title:       item.S(ns, "title"),
		Id:          item.S(ns, "id"),
		PubDate:     item.S(ns, "updated"),
		Description: item.S(ns, "summary"),
	}

	for _, v := range item.SelectNodes(ns, "link") {
		if v.As(ns, "rel") == "enclosure" {
			i.Enclosures = append(i.Enclosures, Enclosure{
				Url:  v.As("", "href"),
				Type: v.As("", "type"),
			})
		} else {
			i.Links = append(i.Links, Link{
				Href:     v.As("", "href"),
				Rel:      v.As("", "rel"),
				Type:     v.As("", "type"),
				HrefLang: v.As("", "hreflang"),
			})
		}
	}

	for _, v := range item.SelectNodes(ns, "contributor") {
		i.Contributors = append(i.Contributors, v.S("", "name"))
	}

	for _, cv := range item.SelectNodes(ns, "category") {
		i.Categories = append(i.Categories, Category{
			Domain: "",
			Text:   cv.As("", "term"),
		})
	}

	if tn := item.SelectNode(ns, "content"); tn != nil {
		i.Content = &Content{
			Type: tn.As("", "type"),
			Lang: tn.S("xml", "lang"),
			Base: tn.S("xml", "base"),
			Text: tn.GetValue(),
		}
	}

	if tn := item.SelectNode(ns, "author"); tn != nil {
		i.Author = Author{
			Name:  tn.S(ns, "name"),
			Uri:   tn.S(ns, "uri"),
			Email: tn.S(ns, "email"),
		}
	}

	return i
}
