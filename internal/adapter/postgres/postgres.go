package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPingTimeout = 5 * time.Second

// PoolConfig holds parameters for the PostgreSQL connection pool.
type PoolConfig struct {
	DatabaseURL       string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	PingTimeout       time.Duration
}

// NewPool initializes a PostgreSQL connection pool using PoolConfig.
func NewPool(ctx context.Context, config PoolConfig) (*pgxpool.Pool, error) {
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("database url cannot be null")
	}

	cfg, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing database configuration: %w", err)
	}

	if config.MaxConns > 0 {
		cfg.MaxConns = config.MaxConns
	}
	if config.MinConns > 0 {
		cfg.MinConns = config.MinConns
	}
	if config.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = config.MaxConnLifetime
	}
	if config.MaxConnIdleTime > 0 {
		cfg.MaxConnIdleTime = config.MaxConnIdleTime
	}
	if config.HealthCheckPeriod > 0 {
		cfg.HealthCheckPeriod = config.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed creating connection pool: %w", err)
	}

	pingTimeout := config.PingTimeout
	if pingTimeout <= 0 {
		pingTimeout = defaultPingTimeout
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed pinging postgres: %w", err)
	}

	return pool, nil
}
