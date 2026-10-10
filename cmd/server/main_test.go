package main

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/hotel-booking/internal/adapter/httpapi/handler"
	"github.com/netologist/hotel-booking/internal/config"
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

func TestRunServer_GracefulShutdown(t *testing.T) {
	port := getFreePort(t)

	cfg := &config.Config{
		HTTPPort:                port,
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

	// Wait briefly for server to bind and listen
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

func TestRunServer_ListenError(t *testing.T) {
	cfg := &config.Config{
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
}

func TestRun_ConfigError(t *testing.T) {
	t.Setenv("APP_ENV", "invalid_environment")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/db")

	err := run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestRun_DatabaseError(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://nonexistent:12345/db?connect_timeout=1")
	t.Setenv("DB_PING_TIMEOUT", "100ms")

	err := run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect postgres")
}
