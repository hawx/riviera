// Riviera is a feed aggregator.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"hawx.me/code/indieauth"
	"hawx.me/code/indieauth/sessions"
	"hawx.me/code/riviera/data"
	"hawx.me/code/riviera/garden"
	"hawx.me/code/riviera/river"
	"hawx.me/code/riviera/river/mapping"
	"hawx.me/code/riviera/subscriptions"
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
      Serve at given socket, instead.`)
}

var (
	cutOff  = flag.String("cutoff", "-24h", "")
	refresh = flag.String("refresh", "15m", "")

	dbPath = flag.String("db", "", "")

	url    = flag.String("url", "http://localhost:8080", "")
	secret = flag.String("secret", "GpgGqpnfFkpjgXj7u3RCdKkoOf/tQqbHkOuuys90Ds4=", "")
	me     = flag.String("me", "", "")

	webPath = flag.String("web", "web", "")
	port    = flag.String("port", "8080", "")
	socket  = flag.String("socket", "", "")
)

func addSubs(from interface{ List() ([]string, error) }, to interface{ Add(string) error }) error {
	subs, err := from.List()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if err := to.Add(sub); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func main() {
	flag.Usage = func() { printHelp() }
	flag.Parse()

	if flag.NArg() == 0 {
		printHelp()
		return
	}

	auth, err := indieauth.Authentication(*url, *url+"/callback")
	if err != nil {
		log.Println(err)
		return
	}

	session, err := sessions.New(*me, *secret, auth)
	if err != nil {
		log.Println(err)
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

	log.Println("opening db at", *dbPath)
	db, err := data.Open(*dbPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer waitFor("db", db.Close)

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

	feeds := river.New(db, river.Options{
		Mapping:   mapping.DefaultMapping,
		CutOff:    duration,
		Refresh:   cacheTimeout,
		LogLength: 500,
	})
	defer waitFor("feeds", feeds.Close)

	riverSubs := db.Subscriptions("river")
	if err = addSubs(riverSubs, feeds); err != nil {
		log.Println(err)
		return
	}

	garden := garden.New(db, garden.Options{})
	defer waitFor("garden", garden.Close)

	gardenSubs := db.Subscriptions("garden")
	if err = addSubs(gardenSubs, garden); err != nil {
		log.Println(err)
		return
	}

	templates, err := template.New("").Funcs(map[string]interface{}{
		"ago": func(t time.Time) string {
			dur := time.Now().Sub(t)
			if dur < time.Minute {
				return fmt.Sprintf("%vs", math.Ceil(dur.Seconds()))
			}
			if dur < time.Hour {
				return fmt.Sprintf("%vm", math.Ceil(dur.Minutes()))
			}
			if dur < 24*time.Hour {
				return fmt.Sprintf("%vh", math.Ceil(dur.Hours()))
			}
			if dur < 31*24*time.Hour {
				return fmt.Sprintf("%vd", math.Ceil(dur.Hours()/24))
			}
			if dur < 365*24*time.Hour {
				return fmt.Sprintf("%vM", math.Ceil(dur.Hours()/24/31))
			}

			return fmt.Sprintf("%vY", math.Ceil(dur.Hours()/24/365))
		},
		"truncate": func(line string, length int) string {
			if len(line) < length {
				return line
			}

			words := strings.Fields(line)
			line = ""
			for _, word := range words {
				if len(line+word) < length {
					line += word + " "
				} else {
					break
				}
			}

			return line[:len(line)-1] + "…"
		},
	}).ParseGlob(*webPath + "/template/*.gotmpl")
	if err != nil {
		fmt.Println(err)
		return
	}

	http.Handle("/", http.RedirectHandler("/river", http.StatusFound))

	http.Handle("/public/", http.StripPrefix("/public",
		http.FileServer(http.Dir(*webPath+"/static"))))

	http.HandleFunc("/river", session.Choose(
		feeds.Handler(templates, true),
		feeds.Handler(templates, false)))

	http.HandleFunc("/garden", session.Choose(
		garden.Handler(templates, true),
		garden.Handler(templates, false)))

	http.HandleFunc("/remove", session.Shield(
		subscriptions.RemoveHandler(subscriptions.Map{
			"river":  {riverSubs, feeds},
			"garden": {gardenSubs, garden},
		}),
	))

	http.HandleFunc("/add", session.Shield(
		subscriptions.AddHandler(subscriptions.Map{
			"river":  {riverSubs, feeds},
			"garden": {gardenSubs, garden},
		}),
	))

	http.HandleFunc("/sign-in", session.SignIn())
	http.HandleFunc("/callback", session.Callback())
	http.HandleFunc("/sign-out", session.SignOut())

	serve.Serve(*port, *socket, http.DefaultServeMux)
	wg.Wait()
}
