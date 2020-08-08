package subscriptions

import (
	"io"
	"log"
	"net/http"
)

type ExecuteTemplate interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

type Map map[string][]interface {
	Add(string) error
	Remove(string) error
}

func RemoveHandler(subsMap Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		where := r.FormValue("where")
		uri := r.FormValue("url")

		subs, ok := subsMap[where]
		if !ok {
			return
		}

		for _, sub := range subs {
			if err := sub.Remove(uri); err != nil {
				log.Println(err)
			}
		}
		log.Println("unsubscribed from", uri, "for", where)

		http.Redirect(w, r, "/"+where, http.StatusFound)
	}
}

func AddHandler(subsMap Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		where := r.FormValue("where")
		uri := r.FormValue("url")

		subs, ok := subsMap[where]
		if !ok {
			return
		}

		for _, sub := range subs {
			if err := sub.Add(uri); err != nil {
				log.Println(err)
			}
		}
		log.Println("subscribed to", uri, "for", where)

		http.Redirect(w, r, "/"+where, http.StatusFound)
	}
}
