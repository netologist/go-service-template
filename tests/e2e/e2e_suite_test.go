package e2e_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	baseURL    string
	httpClient *http.Client
)

const (
	defaultBaseURL      = "http://localhost:8080"
	startupTimeout      = 30 * time.Second
	startupPollInterval = 500 * time.Millisecond
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hotel Booking E2E Test Suite")
}

var _ = BeforeSuite(func() {
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	Eventually(func() error {
		resp, err := httpClient.Get(baseURL + "/healthz")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("readiness probe returned status %d", resp.StatusCode)
		}
		return nil
	}).WithTimeout(startupTimeout).WithPolling(startupPollInterval).Should(
		Succeed(),
		"Target server %s did not become healthy within %s",
		baseURL,
		startupTimeout,
	)
})
