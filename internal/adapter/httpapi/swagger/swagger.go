// Package swagger provides embedded Swagger UI assets and the OpenAPI document handler.
package swagger

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

// assetsFS holds a pinned swagger-ui-dist release (see assets/VERSION.txt).
// Everything is served from the binary: no CDN, no external runtime assets.
//
//go:embed assets
var assetsFS embed.FS

const (
	docsPath = "/docs/"

	// docsCSP restricts the Swagger UI page to same-origin resources. Swagger UI injects inline styles,
	// hence style-src 'unsafe-inline'; scripts stay strictly same-origin.
	docsCSP = "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'"

	docsCacheControl = "public, max-age=3600"
)

// Setup registers the OpenAPI document and the embedded Swagger UI on r.
func Setup(r chi.Router) {
	assets, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic("swagger: embedded assets are missing: " + err.Error())
	}
	files := http.StripPrefix(docsPath, http.FileServerFS(assets))

	r.Get("/openapi.json", func(w http.ResponseWriter, req *http.Request) {
		spec, err := openapi.GetSpec()
		if err != nil {
			slog.ErrorContext(req.Context(), "load openapi spec", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(spec); err != nil {
			slog.ErrorContext(req.Context(), "encode openapi spec", "error", err)
		}
	})

	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, docsPath, http.StatusMovedPermanently)
	})

	r.Get(docsPath+"*", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Security-Policy", docsCSP)
		w.Header().Set("Cache-Control", docsCacheControl)
		files.ServeHTTP(w, req)
	})
}
