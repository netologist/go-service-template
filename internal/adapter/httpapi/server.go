package httpapi

import (
	"github.com/netologist/hotel-booking/internal/adapter/httpapi/handler"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
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
