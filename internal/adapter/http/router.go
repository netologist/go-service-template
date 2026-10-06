package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/netologist/hotel-booking/internal/adapter/http/generated"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
)

const (
	TIMEOUT_DURATION = 60 * time.Second
	MAX_AGE          = 300
)

// NewRouter constructs a Chi HTTP router with standard middlewares and domain endpoints.
func NewRouter(server generated.ServerInterface) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(TIMEOUT_DURATION))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           MAX_AGE,
	}))

	handler.SwaggerRegistry(r)
	generated.HandlerFromMux(server, r)

	return r
}
