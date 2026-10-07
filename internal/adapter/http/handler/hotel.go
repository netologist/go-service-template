package handler

import (
	"encoding/json"
	"net/http"

	v10 "github.com/go-playground/validator/v10"
	"github.com/netologist/hotel-booking/internal/adapter/http/util"
	"github.com/netologist/hotel-booking/internal/adapter/validator"
	"github.com/netologist/hotel-booking/internal/gen/openapi"
	"github.com/netologist/hotel-booking/internal/usecase"
)

func NewHotelHandler(hotelService usecase.HotelService, validator *v10.Validate) *HotelHandler {
	return &HotelHandler{
		hotelService: hotelService,
		validator:    validator,
	}
}

type HotelHandler struct {
	hotelService usecase.HotelService
	validator    *v10.Validate
}

func (h *HotelHandler) CreateHotel(w http.ResponseWriter, r *http.Request) {
	dto, err := h.parseCreateHotelRequest(r)
	if err != nil {
		util.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "malformed JSON body"})
		return
	}

	if err := h.validateCreateHotelRequest(dto); err != nil {
		util.RespondJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation_failed",
			"fields": validator.FormatValidationErrors(err),
		})
		return
	}

	model, err := h.hotelService.CreateHotel(r.Context(), toCreateHotelInput(dto))
	if err != nil {
		util.RespondError(w, err)
		return
	}

	util.RespondJSON(w, http.StatusCreated, toHotelResponse(model))
}

func (h *HotelHandler) FindHotels(w http.ResponseWriter, r *http.Request, params openapi.FindHotelsParams) {
	model, err := h.hotelService.FindHotels(r.Context())
	if err != nil {
		util.RespondError(w, err)
		return
	}

	util.RespondJSON(w, http.StatusOK, toHotelListResponse(model))
}

// helper methods

func (h *HotelHandler) parseCreateHotelRequest(r *http.Request) (openapi.CreateHotelRequest, error) {
	var dto openapi.CreateHotelRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return dto, err
	}
	return dto, nil
}

func (h *HotelHandler) validateCreateHotelRequest(dto openapi.CreateHotelRequest) error {
	return h.validator.Struct(dto)
}

// mapper methods

func toCreateHotelInput(r openapi.CreateHotelRequest) usecase.CreateHotelInput {
	return usecase.CreateHotelInput{
		Name:    r.Name,
		Address: r.Address,
		City:    r.City,
	}
}

func toHotelResponse(output usecase.CreateHotelOutput) openapi.Hotel {
	return openapi.Hotel{
		Id:        output.ID,
		Name:      output.Name,
		City:      output.City,
		Address:   output.Address,
		CreatedAt: output.CreatedAt,
		UpdatedAt: output.UpdatedAt,
	}
}

func toHotelListResponse(hotels []usecase.HotelOutput) []openapi.Hotel {
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
