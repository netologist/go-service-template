package problem

import (
	"encoding/json"
	"net/http"

	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

const (
	ContentTypeProblemJSON = "application/problem+json"
	TypeAboutBlank         = "about:blank"
)

// New creates an openapi.Problem with RFC 9457 defaults.
func New(status int, title, detail, instance, code string, errs []openapi.ValidationError) openapi.Problem {
	p := openapi.Problem{
		Type:   TypeAboutBlank,
		Title:  title,
		Status: status,
	}
	if detail != "" {
		p.Detail = &detail
	}
	if instance != "" {
		p.Instance = &instance
	}
	if code != "" {
		p.Code = &code
	}
	if len(errs) > 0 {
		p.Errors = &errs
	}
	return p
}

// BadRequest creates a 400 Problem.
func BadRequest(detail, instance, code string, errs []openapi.ValidationError) openapi.Problem {
	return New(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), detail, instance, code, errs)
}

// Conflict creates a 409 Problem.
func Conflict(detail, instance, code string) openapi.Problem {
	return New(http.StatusConflict, http.StatusText(http.StatusConflict), detail, instance, code, nil)
}

// InternalServerError creates a 500 Problem.
func InternalServerError(detail, instance string) openapi.Problem {
	return New(
		http.StatusInternalServerError,
		http.StatusText(http.StatusInternalServerError),
		detail,
		instance,
		"INTERNAL_SERVER_ERROR",
		nil,
	)
}

// ServiceUnavailable creates a 503 Problem.
func ServiceUnavailable(detail, instance string) openapi.Problem {
	return New(
		http.StatusServiceUnavailable,
		http.StatusText(http.StatusServiceUnavailable),
		detail,
		instance,
		"SERVICE_UNAVAILABLE",
		nil,
	)
}

// Respond writes an RFC 9457 problem details JSON response.
func Respond(w http.ResponseWriter, p openapi.Problem) {
	w.Header().Set("Content-Type", ContentTypeProblemJSON)
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}
