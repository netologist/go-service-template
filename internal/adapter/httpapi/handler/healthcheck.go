package handler

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/netologist/hotel-booking/internal/adapter/httpapi/problem"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

// HealthCheck handles health and readiness probe requests.
type HealthCheck struct {
	ready atomic.Bool
}

// NewHealthCheck creates a HealthCheck handler initialized to ready state.
func NewHealthCheck() *HealthCheck {
	h := &HealthCheck{}
	h.ready.Store(true)
	return h
}

// SetReady sets the readiness probe state.
func (h *HealthCheck) SetReady(ready bool) {
	h.ready.Store(ready)
}

// IsReady returns the current readiness probe state.
func (h *HealthCheck) IsReady() bool {
	return h.ready.Load()
}

// GetHealth handles GET /healthz. Returns 200 OK when ready, 503 Service Unavailable when draining/unready.
func (h *HealthCheck) GetHealth(
	ctx context.Context,
	request openapi.GetHealthRequestObject,
) (openapi.GetHealthResponseObject, error) {
	if !h.ready.Load() {
		prob := problem.ServiceUnavailable("service is shutting down", "/healthz")
		return HealthCheckUnavailableResponse{problem: prob}, nil
	}

	status := "ok"
	now := time.Now().UTC()
	return openapi.GetHealth200JSONResponse{
		Status:    &status,
		Timestamp: &now,
	}, nil
}

// HealthCheckUnavailableResponse returns an RFC 9457 Problem Details response for unready probes.
type HealthCheckUnavailableResponse struct {
	problem openapi.Problem
}

func (r HealthCheckUnavailableResponse) VisitGetHealthResponse(w http.ResponseWriter) error {
	problem.Respond(w, r.problem)
	return nil
}

var _ openapi.GetHealthResponseObject = (*HealthCheckUnavailableResponse)(nil)
