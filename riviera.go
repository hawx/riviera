package main

import (
	"github.com/hawx/riviera/opml"
	"github.com/hawx/riviera/river"
	"github.com/hawx/riviera/river/database"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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
		"    --cutoff <dur>     # Time to ignore items after (default: -24h)\n",
		"    --refresh <dur>    # Time to refresh feeds after (default: 10m)\n",
		"    --port <num>       # Port to bind to (default: 8080)\n",
		"    --socket <path>    # Serve using unix socket instead\n",
		"\n",
		"    --help             # Display help message\n",
	)
}

var (
	opmlPath = flag.String("opml", "", "")
	dbPath   = flag.String("db", "./db", "")

	cutOff   = flag.String("cutoff", "-24h", "")
	refresh  = flag.String("refresh", "10m", "")
	port     = flag.String("port", "8080", "")
	socket   = flag.String("socket", "", "")

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

	cacheTimeout, err := time.ParseDuration(*refresh)
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
	defer store.Close()

	urls := []string{}
	for _, outline := range subs.Body.Outline {
		urls = append(urls, outline.XmlUrl)
	}

	feeds := river.New(store, duration, cacheTimeout, urls)

	http.HandleFunc("/river.js", func(w http.ResponseWriter, r *http.Request) {
		callback := r.FormValue("callback")
		if callback == "" {
			callback = "onGetRiverStream"
		}

		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(%s)", callback, feeds.Build())
	})

	http.HandleFunc("/-/list", func(w http.ResponseWriter, r *http.Request) {
		list := []string{}
		for _, outline := range subs.Body.Outline {
			list = append(list, outline.XmlUrl)
		}
		data, _ := json.Marshal(list)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(data))
	})

	http.HandleFunc("/-/subscribe", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		feeds.Add(url)
		subs.Body.Outline = append(subs.Body.Outline, opml.Outline{XmlUrl: url})
		subs.Save(*opmlPath)
		w.WriteHeader(204)
	})

	http.HandleFunc("/-/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")

		body := []opml.Outline{}
		for _, outline := range subs.Body.Outline {
			if outline.XmlUrl != url {
				body = append(body, outline)
			}
		}
		subs.Body.Outline = body
		subs.Save(*opmlPath)

		if feeds.Remove(url) {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(400)
		}
	})

	if *socket == "" {
		go func() {
			log.Println("listening on port :" + *port)
			log.Fatal(http.ListenAndServe(":"+*port, nil))
		}()
	} else {
		l, err := net.Listen("unix", *socket)
		if err != nil {
			log.Fatal(err)
		}

		defer l.Close()

		go func() {
			log.Println("listening on", *socket)
			log.Fatal(http.Serve(l, nil))
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s := <-c
	log.Printf("caught %s: shutting down", s)
}
