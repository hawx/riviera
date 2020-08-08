package garden

import (
	"io"
	"log"
	"net/http"

	"hawx.me/code/riviera/garden/gardenjs"
)

type ExecuteTemplate interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

func (garden *Garden) Handler(templates ExecuteTemplate, signedIn bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latest, err := garden.Latest()
		if err != nil {
			log.Println("/garden:", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		type gardenCtx struct {
			Feeds    []gardenjs.Feed
			Page     string
			SignedIn bool
		}

		if err := templates.ExecuteTemplate(w, "garden.gotmpl", gardenCtx{
			Feeds:    latest.Feeds,
			Page:     "garden",
			SignedIn: signedIn,
		}); err != nil {
			log.Println("/garden:", err)
		}
	}
}
