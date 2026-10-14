package main

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/internal/adapter/httpapi/handler"
	"github.com/netologist/go-service-template/internal/config"
)

func getFreePort(t *testing.T) int {
	t.Helper()
	var lc net.ListenConfig
	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	return port
}

func TestRunServer_GracefulShutdown_DualMode(t *testing.T) {
	httpPort := getFreePort(t)
	grpcPort := getFreePort(t)

	cfg := &config.Config{
		ServerMode:              config.ServerModeAll,
		HTTPPort:                httpPort,
		GRPCPort:                grpcPort,
		ServerShutdownTimeout:   2 * time.Second,
		ServerDrainTimeout:      30 * time.Millisecond,
		ServerReadHeaderTimeout: 5 * time.Second,
		ServerReadTimeout:       5 * time.Second,
		ServerWriteTimeout:      5 * time.Second,
		ServerIdleTimeout:       5 * time.Second,
	}

	healthCheck := handler.NewHealthCheck()
	assert.True(t, healthCheck.IsReady())

	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServerWithHealth(ctx, nil, cfg, healthCheck)
	}()

	// Wait briefly for servers to bind and listen
	time.Sleep(50 * time.Millisecond)

	// Trigger graceful shutdown
	cancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
		assert.False(t, healthCheck.IsReady(), "readiness probe should be flipped to false on shutdown")
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down within deadline")
	}
}

func TestRunServer_GracefulShutdown_GRPCOnly(t *testing.T) {
	grpcPort := getFreePort(t)

	cfg := &config.Config{
		ServerMode:            config.ServerModeGRPC,
		GRPCPort:              grpcPort,
		ServerShutdownTimeout: 2 * time.Second,
		ServerDrainTimeout:    20 * time.Millisecond,
	}

	healthCheck := handler.NewHealthCheck()
	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServerWithHealth(ctx, nil, cfg, healthCheck)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down within deadline")
	}
}

func TestRunServer_GracefulShutdown_RESTOnly(t *testing.T) {
	httpPort := getFreePort(t)

	cfg := &config.Config{
		ServerMode:              config.ServerModeREST,
		HTTPPort:                httpPort,
		ServerShutdownTimeout:   2 * time.Second,
		ServerDrainTimeout:      20 * time.Millisecond,
		ServerReadHeaderTimeout: 5 * time.Second,
		ServerReadTimeout:       5 * time.Second,
		ServerWriteTimeout:      5 * time.Second,
		ServerIdleTimeout:       5 * time.Second,
	}

	healthCheck := handler.NewHealthCheck()
	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServerWithHealth(ctx, nil, cfg, healthCheck)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down within deadline")
	}
}

func TestRunServer_ListenError(t *testing.T) {
	t.Run("http listen error", func(t *testing.T) {
		cfg := &config.Config{
			ServerMode:              config.ServerModeREST,
			HTTPPort:                -1,
			ServerShutdownTimeout:   1 * time.Second,
			ServerDrainTimeout:      0,
			ServerReadHeaderTimeout: 1 * time.Second,
			ServerReadTimeout:       1 * time.Second,
			ServerWriteTimeout:      1 * time.Second,
			ServerIdleTimeout:       1 * time.Second,
		}

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := runServer(ctx, nil, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server listen error")
	})

	t.Run("grpc listen error", func(t *testing.T) {
		cfg := &config.Config{
			ServerMode:            config.ServerModeGRPC,
			GRPCPort:              -1,
			ServerShutdownTimeout: 1 * time.Second,
			ServerDrainTimeout:    0,
		}

		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		err := runServer(ctx, nil, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "grpc listen error")
	})
}

func TestParseFlags(t *testing.T) {
	t.Run("no flags retains existing config", func(t *testing.T) {
		cfg := &config.Config{ServerMode: config.ServerModeAll}
		err := parseFlags([]string{}, cfg)
		require.NoError(t, err)
		assert.Equal(t, config.ServerModeAll, cfg.ServerMode)
	})

	t.Run("valid server-mode flag grpc", func(t *testing.T) {
		cfg := &config.Config{ServerMode: config.ServerModeAll}
		err := parseFlags([]string{"--server-mode=grpc"}, cfg)
		require.NoError(t, err)
		assert.Equal(t, config.ServerModeGRPC, cfg.ServerMode)
	})

	t.Run("valid server-mode flag rest", func(t *testing.T) {
		cfg := &config.Config{ServerMode: config.ServerModeAll}
		err := parseFlags([]string{"--server-mode=rest"}, cfg)
		require.NoError(t, err)
		assert.Equal(t, config.ServerModeREST, cfg.ServerMode)
	})

	t.Run("invalid server-mode flag", func(t *testing.T) {
		cfg := &config.Config{ServerMode: config.ServerModeAll}
		err := parseFlags([]string{"--server-mode=invalid"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid --server-mode")
	})
}

func TestRun_ConfigError(t *testing.T) {
	t.Setenv("APP_ENV", "invalid_environment")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/db")

	err := runWithArgs([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestRun_FlagsError(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/db")

	err := runWithArgs([]string{"--server-mode=invalid"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse flags")
}

func TestRun_DatabaseError(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://nonexistent:12345/db?connect_timeout=1")
	t.Setenv("DB_PING_TIMEOUT", "100ms")

	err := runWithArgs([]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect postgres")
}
