package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Environment string `env:"APP_ENV" envDefault:"development" validate:"required,oneof=development staging production test"` //nolint:lll // struct tags cannot be split across lines
	HTTPPort    int    `env:"HTTP_PORT" envDefault:"8080" validate:"required,min=1,max=65535"`
	DatabaseURL string `env:"DATABASE_URL" validate:"required"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info" validate:"required,oneof=debug info warn error"`

	// HTTP Server settings (defaults from cmd/server/main.go)
	ServerShutdownTimeout   time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"10s" validate:"required"`
	ServerDrainTimeout      time.Duration `env:"SERVER_DRAIN_TIMEOUT" envDefault:"5s" validate:"min=0"`
	ServerReadHeaderTimeout time.Duration `env:"SERVER_READ_HEADER_TIMEOUT" envDefault:"30s" validate:"required"`
	ServerReadTimeout       time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"30s" validate:"required"`
	ServerWriteTimeout      time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s" validate:"required"`
	ServerIdleTimeout       time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"2m" validate:"required"`

	// PostgreSQL pool settings (defaults from internal/adapter/postgres/postgres.go)
	DBMaxConns          int32         `env:"DB_MAX_CONNS" envDefault:"25" validate:"required,min=1"`
	DBMinConns          int32         `env:"DB_MIN_CONNS" envDefault:"2" validate:"min=0"`
	DBMaxConnLifetime   time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"30m" validate:"required"`
	DBMaxConnIdleTime   time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"5m" validate:"required"`
	DBHealthCheckPeriod time.Duration `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"1m" validate:"required"`
	DBPingTimeout       time.Duration `env:"DB_PING_TIMEOUT" envDefault:"5s" validate:"required"`
}

// ServerAddr returns the listening address formatted as ":<port>".
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

// Load populates configuration from the environment and validates struct constraints.
func Load() (*Config, error) {
	// bind to config
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	// validate config
	validate := NewValidation("env")
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}
