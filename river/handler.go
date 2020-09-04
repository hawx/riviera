package river

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

func List(feeds River, templates *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		river, err := feeds.Latest()
		if err != nil {
			log.Println("/", err)
			return
		}

		if err := templates.ExecuteTemplate(w, "list.gotmpl", river); err != nil {
			log.Println("/", err)
		}
	})
}

func Log(feeds River, templates *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(feeds.Log()); err != nil {
			log.Println("/log:", err)
		}
	})
}
