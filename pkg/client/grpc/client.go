package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	itemv1 "github.com/netologist/go-service-template/pkg/proto/item/v1"
)

// ItemClient aliases the generated ItemServiceClient interface for consumer convenience.
type ItemClient = itemv1.ItemServiceClient

// Client wraps an active gRPC connection and provides access to service clients.
type Client struct {
	conn       *grpc.ClientConn
	itemClient itemv1.ItemServiceClient
}

// NewItemClient dials the target and returns an ItemServiceClient, the underlying ClientConn, and an error.
// If opts is empty, insecure credentials are used by default.
func NewItemClient(target string, opts ...grpc.DialOption) (itemv1.ItemServiceClient, *grpc.ClientConn, error) {
	conn, err := dial(target, opts...)
	if err != nil {
		return nil, nil, err
	}
	return itemv1.NewItemServiceClient(conn), conn, nil
}

// NewItemClientFromConn creates an ItemServiceClient from an existing gRPC connection interface.
func NewItemClientFromConn(cc grpc.ClientConnInterface) itemv1.ItemServiceClient {
	return itemv1.NewItemServiceClient(cc)
}

// NewClient dials target and bundles the connection and clients into a Client instance.
// If opts is empty, insecure credentials are used by default.
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := dial(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       conn,
		itemClient: itemv1.NewItemServiceClient(conn),
	}, nil
}

// Items returns the ItemServiceClient.
func (c *Client) Items() itemv1.ItemServiceClient {
	return c.itemClient
}

// Conn returns the underlying *grpc.ClientConn.
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var finalOpts []grpc.DialOption
	if len(opts) == 0 {
		finalOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	} else {
		finalOpts = opts
	}

	conn, err := grpc.NewClient(target, finalOpts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", target, err)
	}
	return conn, nil
}
