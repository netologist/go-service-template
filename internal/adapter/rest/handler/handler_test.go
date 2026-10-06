package handler_test

import (
	"net/http"
	"testing"

	"github.com/netologist/hotel-booking/internal/adapter/rest"
	"github.com/netologist/hotel-booking/internal/adapter/rest/handler"
)

func SetupTestRouter(t *testing.T) http.Handler {
	server := rest.NewServer(*handler.NewHealthCheck())
	return rest.NewRouter(server)
}
