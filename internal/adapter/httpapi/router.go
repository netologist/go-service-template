package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/netologist/hotel-booking/internal/adapter/httpapi/problem"
	"github.com/netologist/hotel-booking/internal/adapter/httpapi/swagger"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

const (
	requestTimeout        = 60 * time.Second
	corsMaxAge            = 300
	internalServerMessage = "an unexpected error occurred"
)

// RouterOptions customises router middleware and documentation exposure.
type RouterOptions struct {
	// CORSAllowedOrigins lists exact browser origins allowed to access the API.
	// Empty disables CORS handling entirely.
	CORSAllowedOrigins []string

	// DocsEnabled determines whether /docs and /openapi.json are mounted.
	// Production environments must leave this false.
	DocsEnabled bool
}

// NewRouter constructs a Chi HTTP router with standard middlewares, request validation,
// security headers, and domain endpoints using default options (docs enabled, no CORS).
func NewRouter(server openapi.StrictServerInterface) http.Handler {
	return NewRouterWithOptions(server, RouterOptions{DocsEnabled: true})
}

//nolint:funlen // middleware chain + strict handler wiring
func NewRouterWithOptions(
	server openapi.StrictServerInterface,
	opts RouterOptions,
) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(securityHeaders)

	// Narrowed CORS: exact origins, real methods, no wildcards.
	if len(opts.CORSAllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   opts.CORSAllowedOrigins,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowedHeaders:   []string{"Accept", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           corsMaxAge,
		}))
	}

	if opts.DocsEnabled {
		swagger.Setup(r)
	}

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

	strictOptions := openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			prob := problem.BadRequest(err.Error(), r.URL.Path, "BAD_REQUEST", nil)
			problem.Respond(w, prob)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			reqID := middleware.GetReqID(r.Context())
			slog.ErrorContext(r.Context(), "unhandled server error",
				"error", err,
				"request_id", reqID,
				"path", r.URL.Path,
				"method", r.Method,
			)
			prob := problem.WithRequestID(
				problem.InternalServerError(internalServerMessage, r.URL.Path),
				reqID,
			)
			problem.Respond(w, prob)
		},
	}

	strictHandler := openapi.NewStrictHandlerWithOptions(server, nil, strictOptions)

	// Register API endpoints behind the OpenAPI request validator.
	r.Group(func(api chi.Router) {
		api.Use(nethttpmiddleware.OapiRequestValidatorWithOptions(spec, validatorOptions))
		openapi.HandlerFromMux(strictHandler, api)
	})

	return r
}
