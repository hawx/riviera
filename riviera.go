package main

import (
	"github.com/hawx/riviera/opml"
	"github.com/hawx/riviera/river"
	"github.com/hoisie/web"

	"io/ioutil"
	"log"
	"path/filepath"
)

func asset(path string, ctx *web.Context) string {
	content, err := ioutil.ReadFile("assets/" + path)
	if err != nil {
		log.Fatal(err)
	}

	ctx.ContentType(filepath.Ext(path)[1:])
	return string(content)
}

func style(ctx *web.Context, path string) string {
	return asset("css/"+path, ctx)
}

func script(ctx *web.Context, path string) string {
	return asset("js/"+path, ctx)
}

func image(ctx *web.Context, path string) string {
	return asset("images/"+path, ctx)
}

func index(ctx *web.Context) string {
	return asset("index.html", ctx)
}

func getSubscriptions() []string {
	subs, err := opml.Load("subscriptions.xml")
	if err != nil {
		log.Fatal(err)
	}

	urls := []string{}
	for _, outline := range subs.Body.Outline {
		urls = append(urls, outline.XmlUrl)
	}

	return urls
}

func fetchRiver(ctx *web.Context) string {
	callback, ok := ctx.Params["callback"]
	if !ok {
		callback = "onGetRiverStream"
	}

	ctx.ContentType("js")
	return river.Build(callback, getSubscriptions()...)
}

func main() {
	web.Get("/css/(.*.css)", style)
	web.Get("/js/(.*.js)", script)
	web.Get("/images/(.*)", image)
	web.Get("/river.js", fetchRiver)
	web.Get("/?", index)

	web.Run("0.0.0.0:9999")
}
