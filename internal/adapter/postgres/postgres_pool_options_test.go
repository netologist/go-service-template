package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPoolOptions(t *testing.T) {
	opts := defaultPoolOptions()

	assert.Equal(t, int32(25), opts.maxConns)
	assert.Equal(t, int32(2), opts.minConns)
	assert.Equal(t, 30*time.Minute, opts.maxConnLifetime)
	assert.Equal(t, 5*time.Minute, opts.maxConnIdleTime)
	assert.Equal(t, 1*time.Minute, opts.healthCheckPeriod)
	assert.Equal(t, 5*time.Second, opts.postgresPingTimeout)
	assert.Nil(t, opts.databaseURL)
}

func TestPoolOptions_Overrides(t *testing.T) {
	opts := defaultPoolOptions()

	customMaxConns := int32(50)
	customMinConns := int32(5)
	customLifetime := 1 * time.Hour
	customIdleTime := 10 * time.Minute
	customHealthCheck := 2 * time.Minute
	customPing := 3 * time.Second

	options := []PoolOption{
		WithMaxConns(customMaxConns),
		WithMinConns(customMinConns),
		WithMaxConnLifetime(customLifetime),
		WithMaxConnIdleTime(customIdleTime),
		WithHealthCheckPeriod(customHealthCheck),
		WithPingTimeout(customPing),
	}

	for _, opt := range options {
		opt(&opts)
	}

	assert.Equal(t, customMaxConns, opts.maxConns)
	assert.Equal(t, customMinConns, opts.minConns)
	assert.Equal(t, customLifetime, opts.maxConnLifetime)
	assert.Equal(t, customIdleTime, opts.maxConnIdleTime)
	assert.Equal(t, customHealthCheck, opts.healthCheckPeriod)
	assert.Equal(t, customPing, opts.postgresPingTimeout)
}
