package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool_InvalidURL(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, WithDatabaseURL("invalid-url"), WithMaxConns(10))
	require.Error(t, err)
	assert.Nil(t, pool)
	assert.ErrorContains(t, err, "failed parsing database configuration")
}

func TestNewPool_UnreachableHost(t *testing.T) {
	ctx := context.Background()
	unreachableURL := "postgres://127.0.0.1:59999/nonexistent?sslmode=disable"
	pool, err := NewPool(ctx, WithDatabaseURL(unreachableURL), WithPingTimeout(100*time.Millisecond))
	require.Error(t, err)
	assert.Nil(t, pool)
	assert.ErrorContains(t, err, "failed pinging postgres")
}
