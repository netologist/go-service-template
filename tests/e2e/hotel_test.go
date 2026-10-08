package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/netologist/hotel-booking/internal/gen/openapi"
)

const (
	fieldHotelName    = "name"
	fieldHotelAddress = "address"
	fieldHotelCity    = "city"
)

var _ = Describe("Hotels API (/api/v1/hotels)", func() {
	Describe("POST /api/v1/hotels (Register Hotel)", func() {
		It("creates a new hotel and returns HTTP 201 with generated fields", func() {
			hotelName := fmt.Sprintf("E2E Hotel %s", uuid.NewString()[:8])
			payload, err := json.Marshal(map[string]string{
				fieldHotelName:    hotelName,
				fieldHotelAddress: "100 Grand Boulevard",
				fieldHotelCity:    "Metropolis",
			})
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Post(
				baseURL+"/api/v1/hotels",
				"application/json",
				bytes.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

			var created openapi.Hotel
			err = json.NewDecoder(resp.Body).Decode(&created)
			Expect(err).NotTo(HaveOccurred())

			Expect(created.Id).NotTo(Equal(uuid.Nil))
			Expect(created.Name).To(Equal(hotelName))
			Expect(created.Address).To(Equal("100 Grand Boulevard"))
			Expect(created.City).To(Equal("Metropolis"))
			Expect(created.CreatedAt).NotTo(BeNil())
			Expect(created.UpdatedAt).NotTo(BeNil())
		})

		It("returns HTTP 400 when body has validation errors", func() {
			payload := `{"name":"","address":"","city":""}`
			resp, err := httpClient.Post(
				baseURL+"/api/v1/hotels",
				"application/json",
				strings.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/problem+json"))

			var prob openapi.Problem
			err = json.NewDecoder(resp.Body).Decode(&prob)
			Expect(err).NotTo(HaveOccurred())
			Expect(prob.Status).To(Equal(http.StatusBadRequest))
			Expect(prob.Title).To(Equal("Bad Request"))
			Expect(prob.Detail).NotTo(BeNil())
		})

		It("returns HTTP 400 when body is malformed JSON", func() {
			resp, err := httpClient.Post(
				baseURL+"/api/v1/hotels",
				"application/json",
				strings.NewReader(`{invalid`),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/problem+json"))

			var prob openapi.Problem
			err = json.NewDecoder(resp.Body).Decode(&prob)
			Expect(err).NotTo(HaveOccurred())
			Expect(prob.Status).To(Equal(http.StatusBadRequest))
			Expect(prob.Title).To(Equal("Bad Request"))
			Expect(prob.Detail).NotTo(BeNil())
		})

		It("returns HTTP 409 when attempting to register a duplicate hotel name", func() {
			hotelName := fmt.Sprintf("E2E Unique Hotel %s", uuid.NewString()[:8])
			payload, err := json.Marshal(map[string]string{
				fieldHotelName:    hotelName,
				fieldHotelAddress: "200 Duplicate Ave",
				fieldHotelCity:    "Gotham",
			})
			Expect(err).NotTo(HaveOccurred())

			resp1, err := httpClient.Post(
				baseURL+"/api/v1/hotels",
				"application/json",
				bytes.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp1.Body.Close() }()
			Expect(resp1.StatusCode).To(Equal(http.StatusCreated))

			// Second request with same name
			resp2, err := httpClient.Post(
				baseURL+"/api/v1/hotels",
				"application/json",
				bytes.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp2.Body.Close() }()

			Expect(resp2.StatusCode).To(Equal(http.StatusConflict))
			Expect(resp2.Header.Get("Content-Type")).To(ContainSubstring("application/problem+json"))

			var prob openapi.Problem
			err = json.NewDecoder(resp2.Body).Decode(&prob)
			Expect(err).NotTo(HaveOccurred())
			Expect(prob.Status).To(Equal(http.StatusConflict))
			Expect(prob.Title).To(Equal("Conflict"))
			Expect(prob.Code).NotTo(BeNil())
			Expect(*prob.Code).To(Equal("HOTEL_ALREADY_EXISTS"))
		})
	})

	Describe("GET /api/v1/hotels (List Hotels)", func() {
		It("lists registered hotels and includes newly created hotel", func() {
			hotelName := fmt.Sprintf("E2E Listed Hotel %s", uuid.NewString()[:8])
			createPayload, err := json.Marshal(map[string]string{
				fieldHotelName:    hotelName,
				fieldHotelAddress: "500 Ocean Drive",
				fieldHotelCity:    "Miami Beach",
			})
			Expect(err).NotTo(HaveOccurred())

			createResp, err := httpClient.Post(
				baseURL+"/api/v1/hotels",
				"application/json",
				bytes.NewReader(createPayload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = createResp.Body.Close() }()
			Expect(createResp.StatusCode).To(Equal(http.StatusCreated))

			// Query list
			listResp, err := httpClient.Get(baseURL + "/api/v1/hotels")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = listResp.Body.Close() }()

			Expect(listResp.StatusCode).To(Equal(http.StatusOK))
			Expect(listResp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

			var hotels []openapi.Hotel
			err = json.NewDecoder(listResp.Body).Decode(&hotels)
			Expect(err).NotTo(HaveOccurred())
			Expect(hotels).NotTo(BeEmpty())

			var found *openapi.Hotel
			for i := range hotels {
				if hotels[i].Name == hotelName {
					found = &hotels[i]
					break
				}
			}
			Expect(found).NotTo(BeNil(), "expected created hotel to be found in list")
			Expect(found.Address).To(Equal("500 Ocean Drive"))
			Expect(found.City).To(Equal("Miami Beach"))
		})
	})
})
