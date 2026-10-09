package problem_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/adapter/httpapi/problem"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

func TestNew(t *testing.T) {
	errs := []openapi.ValidationError{
		{Name: "field", Reason: "invalid"},
	}
	p := problem.New(http.StatusBadRequest, "Bad Request", "detail text", "/test", "TEST_CODE", errs)

	assert.Equal(t, problem.TypeAboutBlank, p.Type)
	assert.Equal(t, "Bad Request", p.Title)
	assert.Equal(t, http.StatusBadRequest, p.Status)
	require.NotNil(t, p.Detail)
	assert.Equal(t, "detail text", *p.Detail)
	require.NotNil(t, p.Instance)
	assert.Equal(t, "/test", *p.Instance)
	require.NotNil(t, p.Code)
	assert.Equal(t, "TEST_CODE", *p.Code)
	require.NotNil(t, p.Errors)
	assert.Equal(t, errs, *p.Errors)
}

func TestBadRequest(t *testing.T) {
	p := problem.BadRequest("invalid input", "/api/v1/resource", "INVALID_INPUT", nil)

	assert.Equal(t, http.StatusBadRequest, p.Status)
	assert.Equal(t, "Bad Request", p.Title)
	require.NotNil(t, p.Detail)
	assert.Equal(t, "invalid input", *p.Detail)
	require.NotNil(t, p.Instance)
	assert.Equal(t, "/api/v1/resource", *p.Instance)
	require.NotNil(t, p.Code)
	assert.Equal(t, "INVALID_INPUT", *p.Code)
	assert.Nil(t, p.Errors)
}

func TestConflict(t *testing.T) {
	p := problem.Conflict("already exists", "/api/v1/hotels", "HOTEL_ALREADY_EXISTS")

	assert.Equal(t, http.StatusConflict, p.Status)
	assert.Equal(t, "Conflict", p.Title)
	require.NotNil(t, p.Detail)
	assert.Equal(t, "already exists", *p.Detail)
	require.NotNil(t, p.Code)
	assert.Equal(t, "HOTEL_ALREADY_EXISTS", *p.Code)
}

func TestInternalServerError(t *testing.T) {
	p := problem.InternalServerError("unexpected failure", "/api/v1/hotels")

	assert.Equal(t, http.StatusInternalServerError, p.Status)
	assert.Equal(t, "Internal Server Error", p.Title)
	require.NotNil(t, p.Detail)
	assert.Equal(t, "unexpected failure", *p.Detail)
	require.NotNil(t, p.Code)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", *p.Code)
}

func TestServiceUnavailable(t *testing.T) {
	p := problem.ServiceUnavailable("service is shutting down", "/healthz")

	assert.Equal(t, http.StatusServiceUnavailable, p.Status)
	assert.Equal(t, "Service Unavailable", p.Title)
	require.NotNil(t, p.Detail)
	assert.Equal(t, "service is shutting down", *p.Detail)
	require.NotNil(t, p.Instance)
	assert.Equal(t, "/healthz", *p.Instance)
	require.NotNil(t, p.Code)
	assert.Equal(t, "SERVICE_UNAVAILABLE", *p.Code)
}

func TestRespond(t *testing.T) {
	rec := httptest.NewRecorder()
	prob := problem.ServiceUnavailable("draining connections", "/healthz")

	problem.Respond(rec, prob)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, problem.ContentTypeProblemJSON, rec.Header().Get("Content-Type"))

	var decoded openapi.Problem
	err := json.Unmarshal(rec.Body.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, decoded.Status)
	assert.Equal(t, "Service Unavailable", decoded.Title)
	require.NotNil(t, decoded.Detail)
	assert.Equal(t, "draining connections", *decoded.Detail)
}
