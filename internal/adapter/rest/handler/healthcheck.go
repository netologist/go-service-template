package handler

import (
	"net/http"
	"time"

	"github.com/netologist/hotel-booking/internal/adapter/rest/util"
)

type HealthCheck struct {
}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func (s *HealthCheck) GetHealth(w http.ResponseWriter, r *http.Request) {
	util.RespondJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
