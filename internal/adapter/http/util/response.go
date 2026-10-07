package util

import (
	"encoding/json"
	"net/http"
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

	RespondJSON(w, status, errorResponse{
		Error: err.Error(),
		Code:  code,
	})
}
