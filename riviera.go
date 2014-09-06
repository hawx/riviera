package main

import (
	"github.com/hawx/riviera/opml"
	"github.com/hawx/riviera/river"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

func getSubscriptions() []string {
	subs, err := opml.Load(*opmlPath)
	if err != nil {
		log.Fatal(err)
	}

	urls := []string{}
	for _, outline := range subs.Body.Outline {
		urls = append(urls, outline.XmlUrl)
	}

	return urls
}

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

var (
	assetPath = flag.String("assets", ".", "")
	opmlPath = flag.String("opml", "", "")
	cutOff = flag.String("cutoff", "24h", "")
	port = flag.String("port", "8080", "")
	help = flag.Bool("help", false, "")
)

func main() {
	flag.Parse()

	if *opmlPath == "" || *help {
		printHelp()
		os.Exit(0)
	}

	feeds := river.New(getSubscriptions())

	for _, name := range []string{"", "css/", "js/", "images/"} {
		http.Handle("/" + name, http.StripPrefix("/" + name, http.FileServer(http.Dir(path.Join(*assetPath, name)))))
	}

	http.HandleFunc("/river.js", func(w http.ResponseWriter, r *http.Request) {
		callback := r.FormValue("callback")
		if callback == "" {
			callback = "onGetRiverStream"
		}

		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(%s)", callback, river.FromFeeds(feeds.Latest()))
	})

	log.Println("listening on port :" + *port)
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
