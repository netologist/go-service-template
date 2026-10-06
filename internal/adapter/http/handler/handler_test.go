package handler_test

import (
	"net/http"
	"testing"

	httpadapter "github.com/netologist/hotel-booking/internal/adapter/http"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
)

func SetupTestRouter(t *testing.T) http.Handler {
	server := httpadapter.NewServer(*handler.NewHealthCheck())
	return httpadapter.NewRouter(server)
}
