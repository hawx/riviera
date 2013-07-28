package main

import (
	"github.com/hawx/riviera/opml"
	"github.com/hawx/riviera/river"
	"github.com/hoisie/web"

	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func asset(localPath string, ctx *web.Context) string {
	content, err := ioutil.ReadFile(path.Join(assetPath, localPath))
	if err != nil {
		log.Fatal(err)
	}

	ctx.ContentType(path.Ext(localPath)[1:])
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
	subs, err := opml.Load(opmlPath)
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

	duration, err := time.ParseDuration(cutOff)
	if err != nil {
		log.Fatal(err)
	}

	return river.Build(callback, duration, getSubscriptions()...)
}

var assetPath, opmlPath, cutOff string

func printHelp() {
	fmt.Println(
		"Usage: riviera [options]\n",
		"\n",
		"  Riviera is a river of news feed reader\n",
		"\n",
		"    --opml <path>      # Path to opml file containing feeds to read\n",
		"    --assets <path>    # Path to asset files\n",
		"\n",
		"    --cutoff <secs>    # Time to ignore items after (default: 24h)\n",
		"    --bind <host>      # Host to bind to (default: 0.0.0.0)\n",
		"    --port <num>       # Port to bind to (default: 9999)\n",
		"\n",
		"    --help             # Display help message\n",
	)
}

func main() {
	flag.StringVar(&assetPath, "assets", ".", "")
	flag.StringVar(&opmlPath, "opml", "", "")
	flag.StringVar(&cutOff, "cutoff", "24h", "")

	bind := flag.String("bind", "0.0.0.0", "")
	port := flag.String("port", "9999", "")
	help := flag.Bool("help", false, "")

	flag.Parse()

	if opmlPath == "" || *help {
		printHelp()
		os.Exit(0)
	}

	web.Get("/css/(.*.css)", style)
	web.Get("/js/(.*.js)", script)
	web.Get("/images/(.*)", image)
	web.Get("/river.js", fetchRiver)
	web.Get("/?", index)

	web.Run(*bind + ":" + *port)
}
