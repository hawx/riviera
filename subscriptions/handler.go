package subscriptions

import (
	"io"
	"log"
	"net/http"
	"sort"
)

type ExecuteTemplate interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

type Map map[string]interface {
	Add(string) error
	Remove(string) error
}

func Handler(templates ExecuteTemplate, subsMap Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			action := r.FormValue("action")
			where := r.FormValue("where")
			uri := r.FormValue("url")

			subs, ok := subsMap[where]
			if !ok {
				return
			}

			if action == "add" {
				if err := subs.Add(uri); err != nil {
					log.Println(err)
				}
				log.Println("subscribed to", uri, "for", where)
			} else if action == "remove" {
				if err := subs.Remove(uri); err != nil {
					log.Println(err)
				}
				log.Println("unsubscribed from", uri, "for", where)
			}

			http.Redirect(w, r, "/"+where, http.StatusFound)
			return
		}

		var places []string
		for k := range subsMap {
			places = append(places, k)
		}
		sort.Strings(places)

		if err := templates.ExecuteTemplate(w, "admin.gotmpl", struct {
			Places []string
		}{
			Places: places,
		}); err != nil {
			log.Println("/admin:", err)
		}
	}
}
