package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConns            = 25
	minConns            = 2
	maxConnLifetime     = 30 * time.Minute
	maxConnIdleTime     = 5 * time.Minute
	healthCheckPeriod   = 1 * time.Minute
	postgresPingTimeout = 5 * time.Second
)

// NewPool initializes a PostgreSQL connection pool.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing database configuration: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.HealthCheckPeriod = healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed creating connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, postgresPingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed pinging postgres: %w", err)
	}

	return pool, nil
}
