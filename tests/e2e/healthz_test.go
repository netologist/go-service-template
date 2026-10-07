package e2e_test

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheck API (/healthz)", func() {
	It("returns HTTP 200 with status ok and valid RFC3339 timestamp", func() {
		resp, err := httpClient.Get(baseURL + "/healthz")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

		var body struct {
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		Expect(err).NotTo(HaveOccurred())

		Expect(body.Status).To(Equal("ok"))
		Expect(body.Timestamp).NotTo(BeEmpty())

		parsedTime, err := time.Parse(time.RFC3339, body.Timestamp)
		Expect(err).NotTo(HaveOccurred())
		Expect(time.Since(parsedTime)).To(BeNumerically("<", 10*time.Second))
	})

	It("returns HTTP 405 Method Not Allowed for unsupported methods", func() {
		resp, err := httpClient.Post(baseURL+"/healthz", "application/json", nil)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
	})
})
