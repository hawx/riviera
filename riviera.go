package main

import (
	"github.com/hawx/riviera/river"
	"github.com/hawx/riviera/data"
	"github.com/hawx/riviera/subscriptions"
	"github.com/hawx/riviera/subscriptions/opml"

	"github.com/hawx/serve"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func printHelp() {
	fmt.Println(
		"Usage: riviera [options]\n",
		"\n",
		"  Riviera is a river of news feed generator\n",
		"\n",
		"    --opml <path>      # Path to opml file containing feeds to import\n",
		"    --db <path>        # Path to database\n",
		"\n",
		"    --cutoff <dur>     # Time to ignore items after (default: -24h)\n",
		"    --refresh <dur>    # Time to refresh feeds after (default: 15m)\n",
		"    --port <num>       # Port to bind to (default: 8080)\n",
		"    --socket <path>    # Serve using unix socket instead\n",
		"    --with-admin       # Allow access to admin routes, /-\n",
		"\n",
		"    --help             # Display help message\n",
	)
}

var (
	opmlPath = flag.String("opml", "", "")
	dbPath   = flag.String("db", "./db", "")

	cutOff    = flag.String("cutoff", "-24h", "")
	refresh   = flag.String("refresh", "15m", "")
	port      = flag.String("port", "8080", "")
	socket    = flag.String("socket", "", "")
	withAdmin = flag.Bool("with-admin", false, "")

	help = flag.Bool("help", false, "")
)

const DEFAULT_CALLBACK = "onGetRiverStream"

func main() {
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	duration, err := time.ParseDuration(*cutOff)
	if err != nil {
		log.Fatal(err)
	}

	cacheTimeout, err := time.ParseDuration(*refresh)
	if err != nil {
		log.Fatal(err)
	}

	store, err := data.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	subs, err := subscriptions.Open(store)
	if err != nil {
		log.Fatal(err)
	}

	if *opmlPath != "" {
		outline, err := opml.Load(*opmlPath)
		if err != nil {
			log.Fatal(err)
		}

		subs.Import(outline)
		log.Printf("imported %s\n", *opmlPath)
		return
	}

	feeds := river.New(store, subs, duration, cacheTimeout)

	http.HandleFunc("/river.js", func(w http.ResponseWriter, r *http.Request) {
		callback := r.FormValue("callback")
		if callback == "" {
			callback = DEFAULT_CALLBACK
		}

		w.Header().Set("Content-Type", "application/javascript")

		fmt.Fprintf(w, "%s(", callback)
		feeds.WriteTo(w)
		fmt.Fprintf(w, ")")
	})

	if *withAdmin {
		http.HandleFunc("/-/list", func(w http.ResponseWriter, r *http.Request) {
			list := subs.List()
			data, _ := json.Marshal(list)

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, string(data))
		})

		http.HandleFunc("/-/subscribe", func(w http.ResponseWriter, r *http.Request) {
			url := r.FormValue("url")
			subs.Add(url)
			w.WriteHeader(204)
		})

		http.HandleFunc("/-/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
			url := r.FormValue("url")
			subs.Remove(url)
			w.WriteHeader(204)
		})
	}

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
