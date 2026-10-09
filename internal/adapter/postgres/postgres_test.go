package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool_EmptyURL(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, PoolConfig{})
	require.Error(t, err)
	assert.Nil(t, pool)
	assert.ErrorContains(t, err, "database url cannot be null")
}

func TestNewPool_InvalidURL(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, PoolConfig{
		DatabaseURL: "invalid-url",
		MaxConns:    10,
	})
	require.Error(t, err)
	assert.Nil(t, pool)
	assert.ErrorContains(t, err, "failed parsing database configuration")
}

func TestNewPool_UnreachableHost(t *testing.T) {
	ctx := context.Background()
	unreachableURL := "postgres://127.0.0.1:59999/nonexistent?sslmode=disable"
	pool, err := NewPool(ctx, PoolConfig{
		DatabaseURL: unreachableURL,
		PingTimeout: 100 * time.Millisecond,
	})
	require.Error(t, err)
	assert.Nil(t, pool)
	assert.ErrorContains(t, err, "failed pinging postgres")
}
