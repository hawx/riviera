package main

import (
	"flag"
	"fmt"
	"github.com/hawx/riviera/opml"
	"github.com/hawx/riviera/river"
	"github.com/hawx/riviera/river/database"
	"log"
	"net/http"
	"time"
)

func printHelp() {
	fmt.Println(
		"Usage: riviera [options]\n",
		"\n",
		"  Riviera is a river of news feed reader\n",
		"\n",
		"    --opml <path>      # Path to opml file containing feeds to read\n",
		"    --db <path>        # Path to database\n",
		"\n",
		"    --cutoff <secs>    # Time to ignore items after (default: -24h)\n",
		"    --bind <host>      # Host to bind to (default: 0.0.0.0)\n",
		"    --port <num>       # Port to bind to (default: 9999)\n",
		"\n",
		"    --help             # Display help message\n",
	)
}

var (
	dbPath   = flag.String("db", "./db", "")
	opmlPath = flag.String("opml", "", "")
	cutOff   = flag.String("cutoff", "-24h", "")
	port     = flag.String("port", "8080", "")
	help     = flag.Bool("help", false, "")
)

func main() {
	flag.Parse()

	if *opmlPath == "" || *help {
		printHelp()
		return
	}

	duration, err := time.ParseDuration(*cutOff)
	if err != nil {
		log.Fatal(err)
	}

	subs, err := opml.Load(*opmlPath)
	if err != nil {
		log.Fatal(err)
	}

	store, err := database.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}

	urls := []string{}
	for _, outline := range subs.Body.Outline {
		urls = append(urls, outline.XmlUrl)
	}

	feeds := river.New(store, duration, urls)

	http.HandleFunc("/river.js", func(w http.ResponseWriter, r *http.Request) {
		callback := r.FormValue("callback")
		if callback == "" {
			callback = "onGetRiverStream"
		}

		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(%s)", callback, feeds.Build())
	})

	http.HandleFunc("/-/subscribe", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		feeds.Add(url)
		subs.Body.Outline = append(subs.Body.Outline, opml.Outline{XmlUrl: url})
		subs.Save(*opmlPath)
		w.WriteHeader(204)
	})

	http.HandleFunc("/-/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		if feeds.Remove(r.FormValue("url")) {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(400)
		}
	})

	log.Println("listening on port :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
