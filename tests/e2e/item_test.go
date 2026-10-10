package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/netologist/go-service-template/internal/gen/openapi"
)

const (
	fieldItemName        = "name"
	fieldItemDescription = "description"
	fieldItemPrice       = "price"
)

var _ = Describe("Items API (/api/v1/items)", func() {
	Describe("POST /api/v1/items (Register Item)", func() {
		It("creates a new item and returns HTTP 201 with generated fields", func() {
			itemName := "E2E Item " + uuid.NewString()[:8]
			desc := "High quality testing item"
			price := int64(4999)
			payload, err := json.Marshal(map[string]any{
				fieldItemName:        itemName,
				fieldItemDescription: desc,
				fieldItemPrice:       price,
			})
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Post(
				baseURL+"/api/v1/items",
				"application/json",
				bytes.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

			var created openapi.Item
			err = json.NewDecoder(resp.Body).Decode(&created)
			Expect(err).NotTo(HaveOccurred())

			Expect(created.Id).NotTo(Equal(uuid.Nil))
			Expect(created.Name).To(Equal(itemName))
			Expect(created.Description).To(Equal(desc))
			Expect(created.Price).To(Equal(price))
			Expect(created.CreatedAt).NotTo(BeNil())
			Expect(created.UpdatedAt).NotTo(BeNil())
		})

		It("returns HTTP 400 when body has validation errors", func() {
			payload := `{"name":"","description":"","price":-10}`
			resp, err := httpClient.Post(
				baseURL+"/api/v1/items",
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
				baseURL+"/api/v1/items",
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

		It("returns HTTP 409 when attempting to register a duplicate item name", func() {
			itemName := "E2E Unique Item " + uuid.NewString()[:8]
			payload, err := json.Marshal(map[string]any{
				fieldItemName:        itemName,
				fieldItemDescription: "Unique item description",
				fieldItemPrice:       1999,
			})
			Expect(err).NotTo(HaveOccurred())

			resp1, err := httpClient.Post(
				baseURL+"/api/v1/items",
				"application/json",
				bytes.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp1.Body.Close() }()
			Expect(resp1.StatusCode).To(Equal(http.StatusCreated))

			// Second request with same name
			resp2, err := httpClient.Post(
				baseURL+"/api/v1/items",
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
			Expect(*prob.Code).To(Equal("ITEM_ALREADY_EXISTS"))
		})
	})

	Describe("GET /api/v1/items (List Items)", func() {
		It("lists registered items and includes newly created item", func() {
			itemName := "E2E Listed Item " + uuid.NewString()[:8]
			createPayload, err := json.Marshal(map[string]any{
				fieldItemName:        itemName,
				fieldItemDescription: "Listed item description",
				fieldItemPrice:       2500,
			})
			Expect(err).NotTo(HaveOccurred())

			createResp, err := httpClient.Post(
				baseURL+"/api/v1/items",
				"application/json",
				bytes.NewReader(createPayload),
			)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = createResp.Body.Close() }()
			Expect(createResp.StatusCode).To(Equal(http.StatusCreated))

			// Query list
			listResp, err := httpClient.Get(baseURL + "/api/v1/items")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = listResp.Body.Close() }()

			Expect(listResp.StatusCode).To(Equal(http.StatusOK))
			Expect(listResp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

			var items []openapi.Item
			err = json.NewDecoder(listResp.Body).Decode(&items)
			Expect(err).NotTo(HaveOccurred())
			Expect(items).NotTo(BeEmpty())

			var found *openapi.Item
			for i := range items {
				if items[i].Name == itemName {
					found = &items[i]
					break
				}
			}
			Expect(found).NotTo(BeNil(), "expected created item to be found in list")
			Expect(found.Description).To(Equal("Listed item description"))
			Expect(found.Price).To(Equal(int64(2500)))
		})
	})
})
