package main

import (
	"github.com/hawx/riviera/opml"
	"github.com/hawx/riviera/river"

	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type assetHandler struct {
	path string
}

func AssetHandler(dir string) http.Handler {
	return &assetHandler{path: path.Join(assetPath, dir)}
}

// http.Handler
func (h *assetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if strings.HasPrefix(p, "/") { p = p[1:] }

	filePath := h.path
	if len(p) > 0 {
		filePath = path.Join(assetPath, p)
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if len(p) > 0 {
		w.Header().Set("Content-Type", path.Ext(p)[1:])
	} else {
		w.Header().Set("Content-Type", "text/html")
	}

	fmt.Fprintln(w, string(content))
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

var riverJs string
var fetching bool
var lastFetched time.Time

func fetchRiver() {
	var refreshPoint = time.Now().Add(time.Duration(-15) * time.Minute)
	if fetching || lastFetched.After(refreshPoint) {
		log.Println("feeds still fresh")
		return
	}
	fetching = true
	lastFetched = time.Now()

	duration, err := time.ParseDuration(cutOff)
	if err != nil {
		log.Fatal(err)
	}

	riverJs = river.Build(duration, getSubscriptions()...)
	log.Println("fetched feeds")

	fetching = false
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

	port := flag.String("port", "9999", "")
	help := flag.Bool("help", false, "")

	flag.Parse()

	if opmlPath == "" || *help {
		printHelp()
		os.Exit(0)
	}

	log.Println("starting")
	fetchRiver()

	http.Handle("/css/", AssetHandler("css"))
	http.Handle("/js/", AssetHandler("js"))
	http.Handle("/images/", AssetHandler("images"))

	http.HandleFunc("/river.js", func(w http.ResponseWriter, r *http.Request) {
		fetchRiver()

		callback := r.FormValue("callback")
		if callback == "" {
			callback = "onGetRiverStream"
		}

		w.Header().Set("Content-Type", "application/javascript")

		fmt.Fprintf(w, "%s(%s)", callback, riverJs)
	})

	http.Handle("/", AssetHandler("index.html"))

	log.Println("listening on port :" + *port)
	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
