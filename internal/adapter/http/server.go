package http

import (
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

type Server struct {
	handler.HealthCheck
}

func NewServer(
	healthCheck handler.HealthCheck,
) *Server {
	return &Server{
		HealthCheck: healthCheck,
	}
}

var _ openapi.ServerInterface = (*Server)(nil)
