package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
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

func TestHealthCheck_GetHealth(t *testing.T) {
	h := handler.NewHealthCheck()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.GetHealth(rec, req)

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
}

func TestHealthCheck_Router_Integration(t *testing.T) {
	router := SetupTestRouter(t)

	t.Run("GET /healthz returns ok with timestamp", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/healthz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}
