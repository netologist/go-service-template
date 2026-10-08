package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
	customvalidator "github.com/netologist/hotel-booking/internal/adapter/validator"
	"github.com/netologist/hotel-booking/internal/domain"
	usecasemocks "github.com/netologist/hotel-booking/internal/gen/mocks/usecase"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

const (
	testHotelName          = "Grand Budapest Hotel"
	testHotelAddress       = "100 Alpine Way"
	testHotelCity          = "Lutz"
	validCreateHotelJSON   = `{"name":"Grand Budapest Hotel","address":"100 Alpine Way","city":"Lutz"}`
	contentTypeApplication = "application/json"
	fieldValidationFailed  = "validation_failed"
	fieldName              = "name"
	fieldAddress           = "address"
	fieldCity              = "city"
	errInternalServer      = "INTERNAL_SERVER_ERROR"
)

func TestNewHotelHandler(t *testing.T) {
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()

	h := handler.NewHotelHandler(mockService, v)
	require.NotNil(t, h)
}

func TestHotelHandler_CreateHotel_Success(t *testing.T) {
	hotelID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	mockService.EXPECT().CreateHotel(mock.Anything, testHotelName, testHotelAddress, testHotelCity).Return(domain.Hotel{
		ID:        hotelID,
		Name:      testHotelName,
		Address:   testHotelAddress,
		City:      testHotelCity,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}, nil).Once()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/hotels",
		strings.NewReader(validCreateHotelJSON),
	)
	req.Header.Set("Content-Type", contentTypeApplication)
	rec := httptest.NewRecorder()

	h.CreateHotel(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), contentTypeApplication)

	var got openapi.Hotel
	err := json.NewDecoder(rec.Body).Decode(&got)
	require.NoError(t, err)

	assert.Equal(t, hotelID, got.Id)
	assert.Equal(t, testHotelName, got.Name)
	assert.Equal(t, testHotelAddress, got.Address)
	assert.Equal(t, testHotelCity, got.City)
	require.NotNil(t, got.CreatedAt)
	require.NotNil(t, got.UpdatedAt)
	assert.True(t, fixedTime.Equal(*got.CreatedAt))
	assert.True(t, fixedTime.Equal(*got.UpdatedAt))
}

func TestHotelHandler_CreateHotel_MalformedJSON(t *testing.T) {
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/hotels",
		strings.NewReader(`{"name": "invalid-json`),
	)
	rec := httptest.NewRecorder()

	h.CreateHotel(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]string
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "malformed JSON body", body["error"])
}

type validationTestCase struct {
	name          string
	body          openapi.CreateHotelRequest
	expectedField string
}

func missingFieldTestCases() []validationTestCase {
	return []validationTestCase{
		{
			name: "missing_name",
			body: openapi.CreateHotelRequest{
				Name:    "",
				Address: testHotelAddress,
				City:    testHotelCity,
			},
			expectedField: fieldName,
		},
		{
			name: "missing_address",
			body: openapi.CreateHotelRequest{
				Name:    testHotelName,
				Address: "",
				City:    testHotelCity,
			},
			expectedField: fieldAddress,
		},
		{
			name: "missing_city",
			body: openapi.CreateHotelRequest{
				Name:    testHotelName,
				Address: testHotelAddress,
				City:    "",
			},
			expectedField: fieldCity,
		},
	}
}

func boundaryFieldTestCases() []validationTestCase {
	return []validationTestCase{
		{
			name: "name_too_short",
			body: openapi.CreateHotelRequest{
				Name:    "A",
				Address: testHotelAddress,
				City:    testHotelCity,
			},
			expectedField: fieldName,
		},
		{
			name: "name_too_long",
			body: openapi.CreateHotelRequest{
				Name:    strings.Repeat("A", 256),
				Address: testHotelAddress,
				City:    testHotelCity,
			},
			expectedField: fieldName,
		},
		{
			name: "address_too_short",
			body: openapi.CreateHotelRequest{
				Name:    testHotelName,
				Address: "A",
				City:    testHotelCity,
			},
			expectedField: fieldAddress,
		},
		{
			name: "address_too_long",
			body: openapi.CreateHotelRequest{
				Name:    testHotelName,
				Address: strings.Repeat("A", 256),
				City:    testHotelCity,
			},
			expectedField: fieldAddress,
		},
		{
			name: "city_too_short",
			body: openapi.CreateHotelRequest{
				Name:    testHotelName,
				Address: testHotelAddress,
				City:    "A",
			},
			expectedField: fieldCity,
		},
		{
			name: "city_too_long",
			body: openapi.CreateHotelRequest{
				Name:    testHotelName,
				Address: testHotelAddress,
				City:    strings.Repeat("A", 101),
			},
			expectedField: fieldCity,
		},
	}
}

func TestHotelHandler_CreateHotel_ValidationFailure(t *testing.T) {
	testCases := append(missingFieldTestCases(), boundaryFieldTestCases()...)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockService := usecasemocks.NewMockHotelService(t)
			v := customvalidator.NewForJsonTag()
			h := handler.NewHotelHandler(mockService, v)

			payload, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/api/v1/hotels",
				bytes.NewReader(payload),
			)
			rec := httptest.NewRecorder()

			h.CreateHotel(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var resp struct {
				Error  string                       `json:"error"`
				Fields []customvalidator.FieldError `json:"fields"`
			}
			err = json.NewDecoder(rec.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, fieldValidationFailed, resp.Error)
			require.NotEmpty(t, resp.Fields)

			foundField := false
			for _, f := range resp.Fields {
				if f.Field == tt.expectedField {
					foundField = true
					assert.NotEmpty(t, f.Message)
				}
			}
			assert.True(t, foundField, "expected field error for %s", tt.expectedField)
		})
	}
}

func TestHotelHandler_CreateHotel_ServiceError(t *testing.T) {
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	serviceErr := errors.New("database connection failed")
	mockService.EXPECT().
		CreateHotel(mock.Anything, testHotelName, testHotelAddress, testHotelCity).
		Return(domain.Hotel{}, serviceErr).
		Once()
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/hotels",
		strings.NewReader(validCreateHotelJSON),
	)
	rec := httptest.NewRecorder()

	h.CreateHotel(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var body map[string]string
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, serviceErr.Error(), body["error"])
	assert.Equal(t, errInternalServer, body["code"])
}

func TestHotelHandler_CreateHotel_DomainInvariantError(t *testing.T) {
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	domainErr := domain.ErrInvalidHotel
	mockService.EXPECT().
		CreateHotel(mock.Anything, testHotelName, testHotelAddress, testHotelCity).
		Return(domain.Hotel{}, domainErr).
		Once()
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/hotels",
		strings.NewReader(validCreateHotelJSON),
	)
	rec := httptest.NewRecorder()

	h.CreateHotel(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]string
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, domainErr.Error(), body["error"])
	assert.Equal(t, "INVALID_HOTEL", body["code"])
}

func TestHotelHandler_FindHotels_SuccessWithHotels(t *testing.T) {
	hotel1ID := uuid.New()
	hotel2ID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	expectedHotels := []domain.Hotel{
		{
			ID:        hotel1ID,
			Name:      "Hotel One",
			Address:   "123 First St",
			City:      "Miami",
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
		{
			ID:        hotel2ID,
			Name:      "Hotel Two",
			Address:   "456 Second St",
			City:      "Orlando",
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
	}

	mockService.EXPECT().FindHotels(mock.Anything).Return(expectedHotels, nil).Once()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/hotels", nil)
	rec := httptest.NewRecorder()

	h.FindHotels(rec, req, openapi.FindHotelsParams{})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), contentTypeApplication)

	var got []openapi.Hotel
	err := json.NewDecoder(rec.Body).Decode(&got)
	require.NoError(t, err)

	require.Len(t, got, 2)
	assert.Equal(t, hotel1ID, got[0].Id)
	assert.Equal(t, "Hotel One", got[0].Name)
	assert.Equal(t, "123 First St", got[0].Address)
	assert.Equal(t, "Miami", got[0].City)
	require.NotNil(t, got[0].CreatedAt)
	assert.True(t, fixedTime.Equal(*got[0].CreatedAt))

	assert.Equal(t, hotel2ID, got[1].Id)
	assert.Equal(t, "Hotel Two", got[1].Name)
	assert.Equal(t, "456 Second St", got[1].Address)
	assert.Equal(t, "Orlando", got[1].City)
	require.NotNil(t, got[1].CreatedAt)
	assert.True(t, fixedTime.Equal(*got[1].CreatedAt))
}

func TestHotelHandler_FindHotels_EmptyList(t *testing.T) {
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	mockService.EXPECT().FindHotels(mock.Anything).Return([]domain.Hotel{}, nil).Once()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/hotels", nil)
	rec := httptest.NewRecorder()

	h.FindHotels(rec, req, openapi.FindHotelsParams{})

	assert.Equal(t, http.StatusOK, rec.Code)

	var got []openapi.Hotel
	err := json.NewDecoder(rec.Body).Decode(&got)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestHotelHandler_FindHotels_ServiceError(t *testing.T) {
	mockService := usecasemocks.NewMockHotelService(t)
	v := customvalidator.NewForJsonTag()
	h := handler.NewHotelHandler(mockService, v)

	serviceErr := errors.New("failed to retrieve hotels")
	mockService.EXPECT().FindHotels(mock.Anything).Return(nil, serviceErr).Once()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/hotels", nil)
	rec := httptest.NewRecorder()

	h.FindHotels(rec, req, openapi.FindHotelsParams{})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var body map[string]string
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, serviceErr.Error(), body["error"])
	assert.Equal(t, errInternalServer, body["code"])
}

func TestHotelRouter_Integration_CreateHotel(t *testing.T) {
	hotelID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	router, mockService := SetupTestRouterWithMock(t)

	mockService.EXPECT().CreateHotel(mock.Anything, testHotelName, testHotelAddress, testHotelCity).Return(domain.Hotel{
		ID:        hotelID,
		Name:      testHotelName,
		Address:   testHotelAddress,
		City:      testHotelCity,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}, nil).Once()

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/hotels",
		strings.NewReader(validCreateHotelJSON),
	)
	req.Header.Set("Content-Type", contentTypeApplication)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var got openapi.Hotel
	err := json.NewDecoder(rec.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, hotelID, got.Id)
	assert.Equal(t, testHotelName, got.Name)
}

func TestHotelRouter_Integration_FindHotels(t *testing.T) {
	hotelID := uuid.New()
	fixedTime := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	router, mockService := SetupTestRouterWithMock(t)

	mockService.EXPECT().FindHotels(mock.Anything).Return([]domain.Hotel{
		{
			ID:        hotelID,
			Name:      testHotelName,
			Address:   testHotelAddress,
			City:      testHotelCity,
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
	}, nil).Once()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/hotels", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got []openapi.Hotel
	err := json.NewDecoder(rec.Body).Decode(&got)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, hotelID, got[0].Id)
	assert.Equal(t, testHotelName, got[0].Name)
}
