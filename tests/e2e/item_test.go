package e2e_test

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/netologist/go-service-template/pkg/client"
)

var _ = Describe("Items API (/api/v1/items)", func() {
	Describe("POST /api/v1/items (Register Item)", func() {
		It("creates a new item and returns HTTP 201 with generated fields", func() {
			itemName := "E2E Item " + uuid.NewString()[:8]
			desc := "High quality testing item"
			price := int64(4999)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := apiClient.CreateItemWithResponse(ctx, client.CreateItemJSONRequestBody{
				Name:        itemName,
				Description: &desc,
				Price:       price,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode()).To(Equal(http.StatusCreated))
			Expect(resp.ContentType()).To(ContainSubstring("application/json"))
			Expect(resp.JSON201).NotTo(BeNil())

			created := resp.JSON201
			Expect(created.Id).NotTo(Equal(uuid.Nil))
			Expect(created.Name).To(Equal(itemName))
			Expect(created.Description).To(Equal(desc))
			Expect(created.Price).To(Equal(price))
			Expect(created.CreatedAt).NotTo(BeZero())
			Expect(created.UpdatedAt).NotTo(BeZero())
		})

		It("returns HTTP 400 when body has validation errors", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			payload := `{"name":"","description":"","price":-10}`
			resp, err := apiClient.CreateItemWithBodyWithResponse(
				ctx,
				"application/json",
				strings.NewReader(payload),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode()).To(Equal(http.StatusBadRequest))
			Expect(resp.ContentType()).To(ContainSubstring("application/problem+json"))
			Expect(resp.ApplicationproblemJSON400).NotTo(BeNil())

			prob := resp.ApplicationproblemJSON400
			Expect(prob.Status).To(Equal(http.StatusBadRequest))
			Expect(prob.Title).To(Equal("Bad Request"))
			Expect(prob.Detail).NotTo(BeNil())
		})

		It("returns HTTP 400 when body is malformed JSON", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := apiClient.CreateItemWithBodyWithResponse(
				ctx,
				"application/json",
				strings.NewReader(`{invalid`),
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode()).To(Equal(http.StatusBadRequest))
			Expect(resp.ContentType()).To(ContainSubstring("application/problem+json"))
			Expect(resp.ApplicationproblemJSON400).NotTo(BeNil())

			prob := resp.ApplicationproblemJSON400
			Expect(prob.Status).To(Equal(http.StatusBadRequest))
			Expect(prob.Title).To(Equal("Bad Request"))
			Expect(prob.Detail).NotTo(BeNil())
		})

		It("returns HTTP 409 when attempting to register a duplicate item name", func() {
			itemName := "E2E Unique Item " + uuid.NewString()[:8]
			desc := "Unique item description"

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := client.CreateItemJSONRequestBody{
				Name:        itemName,
				Description: &desc,
				Price:       1999,
			}

			resp1, err := apiClient.CreateItemWithResponse(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp1.StatusCode()).To(Equal(http.StatusCreated))

			// Second request with same name
			resp2, err := apiClient.CreateItemWithResponse(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp2.StatusCode()).To(Equal(http.StatusConflict))
			Expect(resp2.ContentType()).To(ContainSubstring("application/problem+json"))
			Expect(resp2.ApplicationproblemJSON409).NotTo(BeNil())

			prob := resp2.ApplicationproblemJSON409
			Expect(prob.Status).To(Equal(http.StatusConflict))
			Expect(prob.Title).To(Equal("Conflict"))
			Expect(prob.Code).NotTo(BeNil())
			Expect(*prob.Code).To(Equal("ITEM_ALREADY_EXISTS"))
		})
	})

	Describe("GET /api/v1/items (List Items)", func() {
		It("lists registered items and includes newly created item", func() {
			itemName := "E2E Listed Item " + uuid.NewString()[:8]
			desc := "Listed item description"

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			createResp, err := apiClient.CreateItemWithResponse(ctx, client.CreateItemJSONRequestBody{
				Name:        itemName,
				Description: &desc,
				Price:       2500,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.StatusCode()).To(Equal(http.StatusCreated))

			// Query list
			listResp, err := apiClient.FindItemsWithResponse(ctx, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(listResp.StatusCode()).To(Equal(http.StatusOK))
			Expect(listResp.ContentType()).To(ContainSubstring("application/json"))
			Expect(listResp.JSON200).NotTo(BeNil())
			Expect(*listResp.JSON200).NotTo(BeEmpty())

			var found *client.Item
			for i := range *listResp.JSON200 {
				if (*listResp.JSON200)[i].Name == itemName {
					found = &(*listResp.JSON200)[i]
					break
				}
			}
			Expect(found).NotTo(BeNil(), "expected created item to be found in list")
			Expect(found.Description).To(Equal("Listed item description"))
			Expect(found.Price).To(Equal(int64(2500)))
		})
	})
})
