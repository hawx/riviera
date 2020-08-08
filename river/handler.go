package river

import (
	"io"
	"log"
	"net/http"

	"hawx.me/code/riviera/river/riverjs"
)

type ExecuteTemplate interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

func (river *river) Handler(templates ExecuteTemplate, signedIn bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latest := river.Latest()

		type riverCtx struct {
			UpdatedFeeds riverjs.Feeds
			Page         string
			SignedIn     bool
		}

		if err := templates.ExecuteTemplate(w, "river.gotmpl", riverCtx{
			UpdatedFeeds: latest.UpdatedFeeds,
			Page:         "river",
			SignedIn:     signedIn,
		}); err != nil {
			log.Println("/river:", err)
		}
	}
}
