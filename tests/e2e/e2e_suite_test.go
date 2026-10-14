package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	itemv1 "github.com/netologist/go-service-template/pkg/proto/item/v1"
)

var (
	baseURL          string
	grpcTarget       string
	httpClient       *http.Client
	grpcConn         *grpc.ClientConn
	grpcItemClient   itemv1.ItemServiceClient
	grpcHealthClient grpc_health_v1.HealthClient
)

const (
	defaultBaseURL      = "http://localhost:8080"
	defaultGRPCTarget   = "localhost:50051"
	startupTimeout      = 30 * time.Second
	startupPollInterval = 500 * time.Millisecond
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service E2E Test Suite")
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

	grpcTarget = os.Getenv("GRPC_TARGET")
	if grpcTarget == "" {
		grpcTarget = defaultGRPCTarget
	}

	var connErr error
	grpcConn, connErr = grpc.NewClient(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(connErr).NotTo(HaveOccurred())

	grpcItemClient = itemv1.NewItemServiceClient(grpcConn)
	grpcHealthClient = grpc_health_v1.NewHealthClient(grpcConn)

	Eventually(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		resp, err := grpcHealthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			return err
		}
		if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			return fmt.Errorf("grpc health check returned %s", resp.Status)
		}
		return nil
	}).WithTimeout(startupTimeout).WithPolling(startupPollInterval).Should(
		Succeed(),
		"Target gRPC server %s did not become healthy within %s",
		grpcTarget,
		startupTimeout,
	)
})

var _ = AfterSuite(func() {
	if grpcConn != nil {
		_ = grpcConn.Close()
	}
})
