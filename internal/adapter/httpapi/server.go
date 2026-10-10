package httpapi

import (
	"github.com/netologist/go-service-template/internal/adapter/httpapi/handler"
	"github.com/netologist/go-service-template/internal/gen/openapi"
)

type Server struct {
	*handler.HealthCheck
	*handler.ItemHandler
}

func NewServer(
	healthCheck *handler.HealthCheck,
	itemHandler *handler.ItemHandler,
) *Server {
	return &Server{
		HealthCheck: healthCheck,
		ItemHandler: itemHandler,
	}
}

var _ openapi.StrictServerInterface = (*Server)(nil)
