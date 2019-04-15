package river

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// DefaultCallback is the name of the callback to use in the jsonp response.
const DefaultCallback = "onGetRiverStream"

type riverHandler struct {
	River
}

// Handler returns a http.Handler that serves the river.
//
//   /        the riverjs feed wrapped in the DefaultCallback
//   /log     the event log from fetching feeds
func Handler(feeds River) http.Handler {
	return riverHandler{feeds}
}

// ServeHTTP satisfies the http.Handler interface.
func (h riverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "GET" {
		w.Header().Set("Accept", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "%s(", DefaultCallback)
		if err := h.Encode(w); err != nil {
			log.Println("/:", err)
		}
		fmt.Fprintf(w, ")")

	case "/log":
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(h.Log()); err != nil {
			log.Println("/log:", err)
		}

	default:
		http.NotFound(w, r)
	}
}
