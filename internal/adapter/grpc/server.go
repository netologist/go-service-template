package grpc

import (
	"context"
	"log/slog"
	"net"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/netologist/go-service-template/internal/adapter/grpc/handler"
	itemv1 "github.com/netologist/go-service-template/internal/gen/grpc/item/v1"
)

// Server coordinates the gRPC server and health check endpoints.
type Server struct {
	grpcServer   *grpc.Server
	healthServer *health.Server
}

// NewServer initializes a new gRPC Server with health, reflection, and default interceptors.
func NewServer(itemHandler *handler.ItemHandler, opts ...grpc.ServerOption) *Server {
	interceptors := []grpc.UnaryServerInterceptor{
		recoveryUnaryInterceptor(),
		loggingUnaryInterceptor(),
	}

	defaultOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),
	}
	defaultOpts = append(defaultOpts, opts...)

	grpcServer := grpc.NewServer(defaultOpts...)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(itemv1.ItemService_ServiceDesc.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	if itemHandler != nil {
		itemv1.RegisterItemServiceServer(grpcServer, itemHandler)
	}

	return &Server{
		grpcServer:   grpcServer,
		healthServer: healthServer,
	}
}

// Serve starts the gRPC server listening on the provided net.Listener.
func (s *Server) Serve(lis net.Listener) error {
	return s.grpcServer.Serve(lis)
}

// GracefulStop attempts to drain and stop the gRPC server gracefully.
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// Stop forcibly terminates the gRPC server immediately.
func (s *Server) Stop() {
	s.grpcServer.Stop()
}

// SetServingStatus sets the health serving status for a specific service or all services ("").
func (s *Server) SetServingStatus(service string, servingStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.healthServer.SetServingStatus(service, servingStatus)
}

// SetReady sets the overall serving status to SERVING (true) or NOT_SERVING (false).
func (s *Server) SetReady(ready bool) {
	st := grpc_health_v1.HealthCheckResponse_SERVING
	if !ready {
		st = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}
	s.healthServer.SetServingStatus("", st)
	s.healthServer.SetServingStatus(itemv1.ItemService_ServiceDesc.ServiceName, st)
}

// GRPCServer returns the underlying grpc.Server instance.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

// HealthServer returns the underlying health.Server instance.
func (s *Server) HealthServer() *health.Server {
	return s.healthServer
}

func loggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		statusCode := status.Code(err)
		if err != nil && statusCode != codes.OK {
			slog.WarnContext(ctx, "gRPC request completed with error",
				"method", info.FullMethod,
				"duration", duration,
				"code", statusCode.String(),
				"error", err,
			)
		} else {
			slog.DebugContext(ctx, "gRPC request completed",
				"method", info.FullMethod,
				"duration", duration,
				"code", statusCode.String(),
			)
		}

		return resp, err
	}
}

func recoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "gRPC server panic recovered",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				err = status.Errorf(codes.Internal, "panic recovered: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}
