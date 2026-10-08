package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/netologist/hotel-booking/internal/adapter/http/problem"
	"github.com/netologist/hotel-booking/internal/adapter/http/util/swagger"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

const (
	TIMEOUT_DURATION = 60 * time.Second
	MAX_AGE          = 300
)

// NewRouter constructs a Chi HTTP router with standard middlewares, request validation, and domain endpoints.
func NewRouter(server openapi.StrictServerInterface) http.Handler {
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

	// OpenAPI Spec validation middleware
	spec, err := openapi.GetSpec()
	if err != nil {
		panic(fmt.Sprintf("failed to load OpenAPI spec: %v", err))
	}
	validatorOptions := &nethttpmiddleware.Options{
		ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
			prob := problem.New(
				statusCode,
				http.StatusText(statusCode),
				message,
				"",
				"VALIDATION_ERROR",
				nil,
			)
			problem.Respond(w, prob)
		},
		SilenceServersWarning: true,
	}
	validatorOptions.Options.AuthenticationFunc = openapi3filter.NoopAuthenticationFunc
	r.Use(nethttpmiddleware.OapiRequestValidatorWithOptions(spec, validatorOptions))

	swagger.Setup(r)
	strictOptions := openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			prob := problem.BadRequest(err.Error(), r.URL.Path, "BAD_REQUEST", nil)
			problem.Respond(w, prob)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			prob := problem.InternalServerError(err.Error(), r.URL.Path)
			problem.Respond(w, prob)
		},
	}

	strictHandler := openapi.NewStrictHandlerWithOptions(server, nil, strictOptions)
	openapi.HandlerFromMux(strictHandler, r)

	return r
}
