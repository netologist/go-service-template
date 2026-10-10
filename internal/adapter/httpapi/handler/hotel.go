package handler

import (
	"context"
	"errors"

	"github.com/netologist/go-service-template/internal/adapter/httpapi/problem"
	"github.com/netologist/go-service-template/internal/domain"
	"github.com/netologist/go-service-template/internal/gen/openapi"
	"github.com/netologist/go-service-template/internal/usecase"
)

func NewHotelHandler(hotelService usecase.HotelService) *HotelHandler {
	return &HotelHandler{
		hotelService: hotelService,
	}
}

type HotelHandler struct {
	hotelService usecase.HotelService
}

func (h *HotelHandler) CreateHotel(
	ctx context.Context,
	request openapi.CreateHotelRequestObject,
) (openapi.CreateHotelResponseObject, error) {
	model, err := h.hotelService.CreateHotel(ctx, request.Body.Name, request.Body.Address, request.Body.City)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidHotel) {
			prob := problem.BadRequest(err.Error(), "/api/v1/hotels", "INVALID_HOTEL", nil)
			return openapi.CreateHotel400ApplicationProblemPlusJSONResponse{
				BadRequestApplicationProblemPlusJSONResponse: openapi.BadRequestApplicationProblemPlusJSONResponse(prob),
			}, nil
		}
		if errors.Is(err, domain.ErrHotelAlreadyExists) {
			prob := problem.Conflict(domain.ErrHotelAlreadyExists.Error(), "/api/v1/hotels", "HOTEL_ALREADY_EXISTS")
			return openapi.CreateHotel409ApplicationProblemPlusJSONResponse{
				ConflictApplicationProblemPlusJSONResponse: openapi.ConflictApplicationProblemPlusJSONResponse(prob),
			}, nil
		}
		return nil, err
	}

	return openapi.CreateHotel201JSONResponse(toHotelResponse(model)), nil
}

func (h *HotelHandler) FindHotels(
	ctx context.Context,
	request openapi.FindHotelsRequestObject,
) (openapi.FindHotelsResponseObject, error) {
	model, err := h.hotelService.FindHotels(ctx)
	if err != nil {
		return nil, err
	}

	return openapi.FindHotels200JSONResponse(toHotelListResponse(model)), nil
}

// mapper methods

func toHotelResponse(output domain.Hotel) openapi.Hotel {
	return openapi.Hotel{
		Id:        output.ID,
		Name:      output.Name,
		City:      output.City,
		Address:   output.Address,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}
}

func toHotelListResponse(hotels []domain.Hotel) []openapi.Hotel {
	res := make([]openapi.Hotel, 0, len(hotels))
	for _, h := range hotels {
		res = append(res, openapi.Hotel{
			Id:        h.ID,
			Name:      h.Name,
			Address:   h.Address,
			City:      h.City,
			CreatedAt: h.CreatedAt,
			UpdatedAt: h.UpdatedAt,
		})
	}
	return res
}
