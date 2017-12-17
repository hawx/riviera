// Riviera is a feed aggregator.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"
	"hawx.me/code/riviera/river"
	"hawx.me/code/riviera/river/data"
	"hawx.me/code/riviera/river/data/boltdata"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/mapping"
	"hawx.me/code/riviera/subscriptions"
	"hawx.me/code/riviera/subscriptions/opml"
	"hawx.me/code/serve"
)

func printHelp() {
	fmt.Println(`Usage: riviera [options] FILE

  Riviera is a feed aggregator. It reads a list of feeds in OPML
  subscription list format (http://dev.opml.org/spec2.html) given
  as FILE, polls these feeds at a customisable interval, and serves
  a riverjs (http://riverjs.org) format document at '/river'.

  A json list of fetch events is served at '/river/log'

  Changes to FILE are watched and will modify the feeds watched, if it
  can be successfully parsed.

 DISPLAY
   --cutoff DUR='-24h'
      Time to ignore items after, given in standard go duration format
      (see http://golang.org/pkg/time/#ParseDuration).

   --refresh DUR='15m'
      Time to refresh feeds after. This is the default used, but if
      advice is given in the feed itself it may be ignored.

 DATA
   By default riviera runs with an in memory database.

   --boltdb PATH
      Use the boltdb file at the given path.

 SERVE
   --port PORT='8080'
      Serve on given port.

   --socket SOCK
      Serve at given socket, instead.
`)
}

var (
	cutOff  = flag.String("cutoff", "-24h", "")
	refresh = flag.String("refresh", "15m", "")

	boltdbPath = flag.String("boltdb", "", "")

	port   = flag.String("port", "8080", "")
	socket = flag.String("socket", "", "")
)

func loadDatastore() (data.Database, error) {
	if *boltdbPath != "" {
		return boltdata.Open(*boltdbPath)
	}

	return memdata.Open(), nil
}

func watchFile(path string, f func()) (io.Closer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watcher, err
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					f()
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Printf("error watching %s: %v", path, err)
				}
			}
		}
	}()

	return watcher, watcher.Add(path)
}

func main() {
	flag.Usage = func() { printHelp() }
	flag.Parse()

	if flag.NArg() == 0 {
		printHelp()
		return
	}

	var wg sync.WaitGroup
	waitFor := func(name string, f func() error) {
		log.Println(name, "waiting to close")
		wg.Add(1)
		if err := f(); err != nil {
			log.Println("waitFor:", err)
		}
		wg.Done()
		log.Println(name, "closed")
	}

	opmlPath := flag.Arg(0)

	duration, err := time.ParseDuration(*cutOff)
	if err != nil {
		log.Println(err)
		return
	}

	cacheTimeout, err := time.ParseDuration(*refresh)
	if err != nil {
		log.Println(err)
		return
	}

	store, err := loadDatastore()
	if err != nil {
		log.Println(err)
		return
	}
	defer waitFor("datastore", store.Close)

	outline, err := opml.Load(opmlPath)
	if err != nil {
		log.Println(err)
		return
	}

	feeds := river.New(store, river.Options{
		Mapping:   mapping.DefaultMapping,
		CutOff:    duration,
		Refresh:   cacheTimeout,
		LogLength: 500,
	})
	defer waitFor("feeds", feeds.Close)

	subs := subscriptions.FromOpml(outline)
	for _, sub := range subs.List() {
		feeds.Add(sub.URI)
	}

	watcher, err := watchFile(opmlPath, func() {
		log.Printf("reading %s\n", opmlPath)
		outline, err := opml.Load(opmlPath)
		if err != nil {
			log.Printf("could not read %s: %s\n", opmlPath, err)
			return
		}

		added, removed := subscriptions.Diff(subs, subscriptions.FromOpml(outline))
		for _, uri := range added {
			feeds.Add(uri)
			subs.Add(uri)
		}
		for _, uri := range removed {
			feeds.Remove(uri)
			subs.Remove(uri)
		}
	})
	if err != nil {
		log.Printf("could not start watching %s: %v\n", opmlPath, err)
	}
	defer waitFor("watcher", watcher.Close)

	http.Handle("/river/", http.StripPrefix("/river", river.Handler(feeds)))

	serve.Serve(*port, *socket, http.DefaultServeMux)
	wg.Wait()
}
