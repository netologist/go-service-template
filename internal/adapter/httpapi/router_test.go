package httpapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/adapter/httpapi"
	"github.com/netologist/hotel-booking/internal/adapter/httpapi/handler"
	usecasemocks "github.com/netologist/hotel-booking/internal/gen/mocks/usecase"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

func newTestRouter(t *testing.T, opts httpapi.RouterOptions) (http.Handler, *usecasemocks.MockHotelService) {
	t.Helper()
	mockService := usecasemocks.NewMockHotelService(t)
	server := httpapi.NewServer(
		handler.NewHealthCheck(),
		handler.NewHotelHandler(mockService),
	)
	return httpapi.NewRouterWithOptions(server, opts), mockService
}

func TestRouter_SecurityHeaders(t *testing.T) {
	t.Parallel()
	router, _ := newTestRouter(t, httpapi.RouterOptions{DocsEnabled: true})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "default-src 'none'; frame-ancestors 'none'", rec.Header().Get("Content-Security-Policy"))
}

func TestRouter_CORS_ConfiguredAndNarrow(t *testing.T) {
	t.Parallel()
	allowedOrigin := "https://admin.example.com"
	router, _ := newTestRouter(t, httpapi.RouterOptions{
		CORSAllowedOrigins: []string{allowedOrigin},
	})

	t.Run("allowed origin preflight succeeds", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, "/api/v1/hotels", nil)
		req.Header.Set("Origin", allowedOrigin)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, allowedOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("disallowed origin is rejected", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, "/api/v1/hotels", nil)
		req.Header.Set("Origin", "https://attacker.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestRouter_DocsGating(t *testing.T) {
	t.Parallel()

	t.Run("docs disabled returns 404", func(t *testing.T) {
		t.Parallel()
		router, _ := newTestRouter(t, httpapi.RouterOptions{DocsEnabled: false})

		reqDocs := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs", nil)
		recDocs := httptest.NewRecorder()
		router.ServeHTTP(recDocs, reqDocs)
		assert.Equal(t, http.StatusNotFound, recDocs.Code)

		reqOpenAPI := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/openapi.json", nil)
		recOpenAPI := httptest.NewRecorder()
		router.ServeHTTP(recOpenAPI, reqOpenAPI)
		assert.Equal(t, http.StatusNotFound, recOpenAPI.Code)
	})

	t.Run("docs enabled serves assets and spec without CDN", func(t *testing.T) {
		t.Parallel()
		router, _ := newTestRouter(t, httpapi.RouterOptions{DocsEnabled: true})

		reqDocs := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs", nil)
		recDocs := httptest.NewRecorder()
		router.ServeHTTP(recDocs, reqDocs)
		assert.Equal(t, http.StatusMovedPermanently, recDocs.Code)

		reqIndex := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs/", nil)
		recIndex := httptest.NewRecorder()
		router.ServeHTTP(recIndex, reqIndex)
		assert.Equal(t, http.StatusOK, recIndex.Code)
		assert.NotContains(t, recIndex.Body.String(), "unpkg.com")
		assert.Contains(t, recIndex.Body.String(), "swagger-ui-bundle.js")

		reqOpenAPI := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/openapi.json", nil)
		recOpenAPI := httptest.NewRecorder()
		router.ServeHTTP(recOpenAPI, reqOpenAPI)
		assert.Equal(t, http.StatusOK, recOpenAPI.Code)
	})
}

func TestRouter_InternalServerError_GenericMessageAndRequestID(t *testing.T) {
	t.Parallel()
	router, mockService := newTestRouter(t, httpapi.RouterOptions{})

	rawDBError := errors.New("pq: password authentication failed for user \"postgres\"")
	mockService.EXPECT().
		FindHotels(mock.Anything).
		Return(nil, rawDBError).
		Once()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/hotels", nil)
	req.Header.Set("X-Request-Id", "client-trace-12345")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var prob openapi.Problem
	err := json.NewDecoder(rec.Body).Decode(&prob)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, prob.Status)
	assert.Equal(t, "Internal Server Error", prob.Title)
	require.NotNil(t, prob.Detail)
	assert.Equal(t, "an unexpected error occurred", *prob.Detail)
	assert.NotContains(t, *prob.Detail, "postgres")
	assert.NotContains(t, *prob.Detail, "password")
	require.NotNil(t, prob.RequestId)
	assert.NotEmpty(t, *prob.RequestId)
}
