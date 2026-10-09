package handler

import (
	"context"
	"time"

	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

type HealthCheck struct{}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func (h *HealthCheck) GetHealth(
	ctx context.Context,
	request openapi.GetHealthRequestObject,
) (openapi.GetHealthResponseObject, error) {
	status := "ok"
	now := time.Now().UTC()
	return openapi.GetHealth200JSONResponse{
		Status:    &status,
		Timestamp: &now,
	}, nil
}
