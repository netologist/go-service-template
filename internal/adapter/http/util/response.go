package util

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/netologist/hotel-booking/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func RespondError(w http.ResponseWriter, err error) {
	var (
		status = http.StatusInternalServerError
		code   = "INTERNAL_SERVER_ERROR"
	)

	if errors.Is(err, domain.ErrInvalidHotel) {
		status = http.StatusBadRequest
		code = "INVALID_HOTEL"
	}

	RespondJSON(w, status, errorResponse{
		Error: err.Error(),
		Code:  code,
	})
}
