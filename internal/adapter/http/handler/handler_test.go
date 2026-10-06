package handler_test

import (
	"net/http"
	"testing"

	httpinfra "github.com/netologist/hotel-booking/internal/adapter/http"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
)

func SetupTestRouter(t *testing.T) http.Handler {
	server := httpinfra.NewServer(*handler.NewHealthCheck())
	return httpinfra.NewRouter(server)
}
