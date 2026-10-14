package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	grpcadapter "github.com/netologist/go-service-template/internal/adapter/grpc"
	"github.com/netologist/go-service-template/internal/adapter/grpc/handler"
	itemv1 "github.com/netologist/go-service-template/internal/gen/grpc/item/v1"
	usecasemocks "github.com/netologist/go-service-template/internal/gen/mocks/usecase"
)

const bufSize = 1024 * 1024

func TestServer_InitializationAndHealth(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	itemHandler := handler.NewItemHandler(mockService)

	srv := grpcadapter.NewServer(itemHandler)
	require.NotNil(t, srv)
	assert.NotNil(t, srv.GRPCServer())
	assert.NotNil(t, srv.HealthServer())

	lis := bufconn.Listen(bufSize)
	defer func() { _ = lis.Close() }()

	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	healthClient := grpc_health_v1.NewHealthClient(conn)

	// Check initially SERVING
	resp, err := healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)

	// Flip ready to false
	srv.SetReady(false)
	resp, err = healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, resp.Status)

	// Flip ready to true
	srv.SetReady(true)
	resp, err = healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
}

func TestServer_RecoveryInterceptor(t *testing.T) {
	mockService := usecasemocks.NewMockItemService(t)
	itemHandler := handler.NewItemHandler(mockService)

	// Panic inside usecase
	mockService.EXPECT().FindItems(mock.Anything).Panic("unexpected fatal crash").Once()

	srv := grpcadapter.NewServer(itemHandler)
	lis := bufconn.Listen(bufSize)
	defer func() { _ = lis.Close() }()

	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := itemv1.NewItemServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.FindItems(ctx, &itemv1.FindItemsRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "panic recovered")
}
