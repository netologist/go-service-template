package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/adapter/httpapi/handler"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

const (
	healthStatusOK          = "ok"
	healthContentTypeHeader = "application/json"
	maxTimestampDrift       = 5 * time.Second
)

func TestNewHealthCheck(t *testing.T) {
	h := handler.NewHealthCheck()
	require.NotNil(t, h)
}

func TestHealthCheck_Readiness(t *testing.T) {
	h := handler.NewHealthCheck()
	assert.True(t, h.IsReady())

	h.SetReady(false)
	assert.False(t, h.IsReady())

	h.SetReady(true)
	assert.True(t, h.IsReady())
}

func TestHealthCheck_GetHealth_WhenNotReady(t *testing.T) {
	h := handler.NewHealthCheck()
	h.SetReady(false)

	resp, err := h.GetHealth(t.Context(), openapi.GetHealthRequestObject{})
	require.NoError(t, err)
	require.IsType(t, handler.HealthCheckUnavailableResponse{}, resp)
}

func TestHealthCheck_GetHealth(t *testing.T) {
	h := handler.NewHealthCheck()

	resp, err := h.GetHealth(t.Context(), openapi.GetHealthRequestObject{})
	require.NoError(t, err)

	healthResp, ok := resp.(openapi.GetHealth200JSONResponse)
	require.True(t, ok)
	require.NotNil(t, healthResp.Status)
	assert.Equal(t, healthStatusOK, *healthResp.Status)
	require.NotNil(t, healthResp.Timestamp)
	assert.WithinDuration(t, time.Now().UTC(), *healthResp.Timestamp, maxTimestampDrift)
}

func TestHealthCheck_Router_Integration(t *testing.T) {
	router := SetupTestRouter(t)

	t.Run("GET /healthz returns ok with timestamp", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), healthContentTypeHeader)

		var body healthResponse
		err := json.NewDecoder(rec.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, healthStatusOK, body.Status)
		require.NotEmpty(t, body.Timestamp)

		parsedTime, err := time.Parse(time.RFC3339, body.Timestamp)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now().UTC(), parsedTime, maxTimestampDrift)
	})

	t.Run("POST /healthz returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/healthz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("GET /healthz returns 503 service unavailable when not ready", func(t *testing.T) {
		health := handler.NewHealthCheck()
		routerWithHealth := SetupTestRouterWithHealth(t, health)
		health.SetReady(false)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		routerWithHealth.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "SERVICE_UNAVAILABLE")
		assert.Contains(t, rec.Body.String(), "service is shutting down")
	})
}
