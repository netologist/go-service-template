package handler_test

import (
	"net/http"
	"testing"

	httpadapter "github.com/netologist/hotel-booking/internal/adapter/http"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
	"github.com/netologist/hotel-booking/internal/adapter/validator"
	usecasemocks "github.com/netologist/hotel-booking/internal/gen/mocks/usecase"
)

func SetupTestRouter(t *testing.T) http.Handler {
	router, _ := SetupTestRouterWithMock(t)
	return router
}

func SetupTestRouterWithMock(t *testing.T) (http.Handler, *usecasemocks.MockHotelService) {
	mockHotelService := usecasemocks.NewMockHotelService(t)
	server := httpadapter.NewServer(
		handler.NewHealthCheck(),
		handler.NewHotelHandler(mockHotelService, validator.NewForJsonTag()),
	)
	return httpadapter.NewRouter(server), mockHotelService
}
