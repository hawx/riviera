package main

import (
	"hawx.me/code/riviera/data"
	"hawx.me/code/riviera/data/boltdata"
	"hawx.me/code/riviera/data/memdata"
	"hawx.me/code/riviera/river"
	"hawx.me/code/riviera/subscriptions"
	"hawx.me/code/riviera/subscriptions/opml"

	"hawx.me/code/serve"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func printHelp() {
	fmt.Println(`Usage: riviera [options]

  Riviera is a river of news feed generator.

   --opml PATH
      Import subscriptions from opml feed, then quit.

 DISPLAY
   --cutoff DUR='-24h'
      Time to ignore items after, in standard go duration format.

   --refresh DUR='15m'
      Time to refresh feeds after.

 DATA
   --boltdb PATH
      Use the boltdb file at the given path.

   --memdb
      Use an in memory database, default.

 SERVE
   --with-admin
      Serve admin routes at '/-'.

   --port PORT='8080'
      Serve on given port.

   --socket SOCK
      Serve at given socket.
`)
}

var (
	opmlPath = flag.String("opml", "", "")

	cutOff  = flag.String("cutoff", "-24h", "")
	refresh = flag.String("refresh", "15m", "")

	boltdbPath = flag.String("boltdb", "", "")
	memdbFlag  = flag.Bool("memdb", true, "")

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

	var store data.Database = memdata.Open()

	if *boltdbPath != "" {
		store, err = boltdata.Open(*boltdbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer store.Close()
	}

	subs, err := subscriptions.Open(store)
	if err != nil {
		log.Fatal(err)
	}

	if *opmlPath != "" {
		outline, err := opml.Load(*opmlPath)
		if err != nil {
			log.Fatal(err)
		}

		subscriptions.FromOpml(subs, outline)
		log.Printf("imported %s\n", *opmlPath)
		return
	}

	feeds := river.New(store, subs, river.Options{
		Mapping: river.DefaultMapping,
		CutOff:  duration,
		Refresh: cacheTimeout,
	})

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

	http.HandleFunc("/subscriptions.opml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		subscriptions.AsOpml(subs).WriteTo(w)
	})

	if *withAdmin {
		http.HandleFunc("/-/list", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(subs.List())
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
