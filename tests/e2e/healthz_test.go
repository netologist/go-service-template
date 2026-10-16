package e2e_test

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheck API (/healthz)", func() {
	It("returns HTTP 200 with status ok and valid RFC3339 timestamp", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := apiClient.GetHealthWithResponse(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.StatusCode()).To(Equal(http.StatusOK))
		Expect(resp.ContentType()).To(ContainSubstring("application/json"))
		Expect(resp.JSON200).NotTo(BeNil())
		Expect(resp.JSON200.Status).NotTo(BeNil())
		Expect(*resp.JSON200.Status).To(Equal("ok"))
		Expect(resp.JSON200.Timestamp).NotTo(BeNil())
		Expect(time.Since(*resp.JSON200.Timestamp)).To(BeNumerically("<", 10*time.Second))
	})

	It("returns HTTP 405 Method Not Allowed for unsupported methods", func() {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/healthz", nil)
		Expect(err).NotTo(HaveOccurred())
		resp, err := httpClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
	})
})
