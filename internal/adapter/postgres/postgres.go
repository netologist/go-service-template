package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes a PostgreSQL connection pool with optional configuration parameters.
func NewPool(ctx context.Context, opts ...PoolOption) (*pgxpool.Pool, error) {
	options := defaultPoolOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if options.databaseURL == "" {
		return nil, fmt.Errorf("database url cannot be null")
	}

	cfg, err := pgxpool.ParseConfig(options.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing database configuration: %w", err)
	}

	cfg.MaxConns = options.maxConns
	cfg.MinConns = options.minConns
	cfg.MaxConnLifetime = options.maxConnLifetime
	cfg.MaxConnIdleTime = options.maxConnIdleTime
	cfg.HealthCheckPeriod = options.healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed creating connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, options.postgresPingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed pinging postgres: %w", err)
	}

	return pool, nil
}
