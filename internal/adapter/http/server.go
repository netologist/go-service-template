package http

import (
	"github.com/netologist/hotel-booking/internal/adapter/http/generated"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
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

// var _ generated.ServerInterface = (*Server)(nil)
var _ generated.ServerInterface = &Server{}
