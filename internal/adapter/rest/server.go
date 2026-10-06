package rest

import (
	api "github.com/netologist/hotel-booking/api/generated"
	"github.com/netologist/hotel-booking/internal/adapter/rest/handler"
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

var _ api.ServerInterface = (*Server)(nil)
