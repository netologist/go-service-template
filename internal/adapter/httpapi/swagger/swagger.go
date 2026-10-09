package swagger

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

//go:embed index.html
var swaggerHTML []byte

func Setup(r chi.Router) {
	r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		spec, err := openapi.GetSpec()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(spec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(swaggerHTML)
	})
}
