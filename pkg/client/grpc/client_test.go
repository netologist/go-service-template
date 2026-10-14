package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	clientgrpc "github.com/netologist/go-service-template/pkg/client/grpc"
	itemv1 "github.com/netologist/go-service-template/pkg/proto/item/v1"
)

const (
	testBufSize   = 1024 * 1024
	testItemName  = "Mechanical Keyboard"
	testItemDesc  = "Ergonomic hot-swappable mechanical keyboard"
	testItemPrice = int64(9900)
)

type stubItemServiceServer struct {
	itemv1.UnimplementedItemServiceServer
	createItemFn func(context.Context, *itemv1.CreateItemRequest) (*itemv1.CreateItemResponse, error)
	findItemsFn  func(context.Context, *itemv1.FindItemsRequest) (*itemv1.FindItemsResponse, error)
}

func (s *stubItemServiceServer) CreateItem(
	ctx context.Context,
	req *itemv1.CreateItemRequest,
) (*itemv1.CreateItemResponse, error) {
	if s.createItemFn != nil {
		return s.createItemFn(ctx, req)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *stubItemServiceServer) FindItems(
	ctx context.Context,
	req *itemv1.FindItemsRequest,
) (*itemv1.FindItemsResponse, error) {
	if s.findItemsFn != nil {
		return s.findItemsFn(ctx, req)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func setupTestServer(t *testing.T, srv *stubItemServiceServer) *bufconn.Listener {
	t.Helper()
	lis := bufconn.Listen(testBufSize)

	server := grpc.NewServer()
	itemv1.RegisterItemServiceServer(server, srv)

	go func() {
		_ = server.Serve(lis)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = lis.Close()
	})

	return lis
}

func bufDialer(lis *bufconn.Listener) grpc.DialOption {
	return grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	})
}

func TestClient_NewItemClient(t *testing.T) {
	itemID := uuid.New().String()
	fixedTime := timestamppb.Now()

	stub := &stubItemServiceServer{
		createItemFn: func(_ context.Context, req *itemv1.CreateItemRequest) (*itemv1.CreateItemResponse, error) {
			return &itemv1.CreateItemResponse{
				Item: &itemv1.Item{
					Id:          itemID,
					Name:        req.GetName(),
					Description: req.GetDescription(),
					Price:       req.GetPrice(),
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				},
			}, nil
		},
	}

	lis := setupTestServer(t, stub)

	dialOpts := []grpc.DialOption{
		bufDialer(lis),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	itemClient, conn, err := clientgrpc.NewItemClient("passthrough://bufnet", dialOpts...)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := itemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
		Name:        testItemName,
		Description: testItemDesc,
		Price:       testItemPrice,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, itemID, resp.Item.Id)
	assert.Equal(t, testItemName, resp.Item.Name)
}

func TestClient_NewClient(t *testing.T) {
	stub := &stubItemServiceServer{
		findItemsFn: func(_ context.Context, _ *itemv1.FindItemsRequest) (*itemv1.FindItemsResponse, error) {
			return &itemv1.FindItemsResponse{
				Items: []*itemv1.Item{
					{
						Id:          uuid.New().String(),
						Name:        testItemName,
						Description: testItemDesc,
						Price:       testItemPrice,
						CreatedAt:   timestamppb.Now(),
						UpdatedAt:   timestamppb.Now(),
					},
				},
			}, nil
		},
	}

	lis := setupTestServer(t, stub)

	dialOpts := []grpc.DialOption{
		bufDialer(lis),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	client, err := clientgrpc.NewClient("passthrough://bufnet", dialOpts...)
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.NotNil(t, client.Conn())
	assert.NotNil(t, client.Items())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Items().FindItems(ctx, &itemv1.FindItemsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, testItemName, resp.Items[0].Name)
}

func TestClient_NewItemClientFromConn(t *testing.T) {
	stub := &stubItemServiceServer{
		createItemFn: func(_ context.Context, req *itemv1.CreateItemRequest) (*itemv1.CreateItemResponse, error) {
			if len(req.GetName()) < 2 {
				return nil, status.Error(codes.InvalidArgument, "invalid item: name too short")
			}
			return &itemv1.CreateItemResponse{}, nil
		},
	}

	lis := setupTestServer(t, stub)

	conn, err := grpc.NewClient("passthrough://bufnet",
		bufDialer(lis),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	itemClient := clientgrpc.NewItemClientFromConn(conn)
	require.NotNil(t, itemClient)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = itemClient.CreateItem(ctx, &itemv1.CreateItemRequest{
		Name:  "A",
		Price: 10,
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
