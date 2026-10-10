package httpapi

import (
	"github.com/netologist/go-service-template/internal/adapter/httpapi/handler"
	"github.com/netologist/go-service-template/internal/gen/openapi"
)

type Server struct {
	*handler.HealthCheck
	*handler.HotelHandler
}

func NewServer(
	healthCheck *handler.HealthCheck,
	hotelHandler *handler.HotelHandler,
) *Server {
	return &Server{
		HealthCheck:  healthCheck,
		HotelHandler: hotelHandler,
	}
}

var _ openapi.StrictServerInterface = (*Server)(nil)
