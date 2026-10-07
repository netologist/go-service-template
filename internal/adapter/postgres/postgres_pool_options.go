package postgres

import (
	"time"
)

const (
	defaultMaxConns            = 25
	defaultMinConns            = 2
	defaultMaxConnLifetime     = 30 * time.Minute
	defaultMaxConnIdleTime     = 5 * time.Minute
	defaultHealthCheckPeriod   = 1 * time.Minute
	defaultPostgresPingTimeout = 5 * time.Second
)

type poolOptions struct {
	maxConns            int32
	minConns            int32
	maxConnLifetime     time.Duration
	maxConnIdleTime     time.Duration
	healthCheckPeriod   time.Duration
	postgresPingTimeout time.Duration
	databaseURL         string
}

func defaultPoolOptions() poolOptions {
	return poolOptions{
		maxConns:            defaultMaxConns,
		minConns:            defaultMinConns,
		maxConnLifetime:     defaultMaxConnLifetime,
		maxConnIdleTime:     defaultMaxConnIdleTime,
		healthCheckPeriod:   defaultHealthCheckPeriod,
		postgresPingTimeout: defaultPostgresPingTimeout,
	}
}

// PoolOption defines a functional option for configuring the connection pool.
type PoolOption func(*poolOptions)

// Option is an alias for PoolOption.
type Option = PoolOption

// WithDatabaseURL sets the databaseUrl of connections in the pool.
func WithDatabaseURL(s string) PoolOption {
	return func(o *poolOptions) {
		o.databaseURL = s
	}
}

// WithMaxConns sets the maximum number of connections in the pool.
func WithMaxConns(n int32) PoolOption {
	return func(o *poolOptions) {
		o.maxConns = n
	}
}

// WithMinConns sets the minimum number of idle connections in the pool.
func WithMinConns(n int32) PoolOption {
	return func(o *poolOptions) {
		o.minConns = n
	}
}

// WithMaxConnLifetime sets the maximum lifetime of connections in the pool.
func WithMaxConnLifetime(d time.Duration) PoolOption {
	return func(o *poolOptions) {
		o.maxConnLifetime = d
	}
}

// WithMaxConnIdleTime sets the maximum idle time of connections in the pool.
func WithMaxConnIdleTime(d time.Duration) PoolOption {
	return func(o *poolOptions) {
		o.maxConnIdleTime = d
	}
}

// WithHealthCheckPeriod sets the duration between health checks on idle connections.
func WithHealthCheckPeriod(d time.Duration) PoolOption {
	return func(o *poolOptions) {
		o.healthCheckPeriod = d
	}
}

// WithPingTimeout sets the timeout for pinging the database during initialization.
func WithPingTimeout(d time.Duration) PoolOption {
	return func(o *poolOptions) {
		o.postgresPingTimeout = d
	}
}
