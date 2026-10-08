package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/netologist/hotel-booking/internal/adapter/validator"
)

type Config struct {
	Environment string `envconfig:"APP_ENV" default:"development" validate:"required,oneof=development staging production test"` //nolint:lll // struct tags cannot be split across lines
	HTTPPort    int    `envconfig:"HTTP_PORT" default:"8080" validate:"required,min=1,max=65535"`
	DatabaseURL string `envconfig:"DATABASE_URL" validate:"required"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=debug info warn error"`

	// HTTP Server settings (defaults from cmd/server/main.go)
	ServerShutdownTimeout   time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"10s" validate:"required"`
	ServerReadHeaderTimeout time.Duration `envconfig:"SERVER_READ_HEADER_TIMEOUT" default:"30s" validate:"required"`
	ServerReadTimeout       time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"30s" validate:"required"`
	ServerWriteTimeout      time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"30s" validate:"required"`
	ServerIdleTimeout       time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"2m" validate:"required"`

	// PostgreSQL pool settings (defaults from internal/adapter/postgres/postgres.go)
	DBMaxConns          int32         `envconfig:"DB_MAX_CONNS" default:"25" validate:"required,min=1"`
	DBMinConns          int32         `envconfig:"DB_MIN_CONNS" default:"2" validate:"min=0"`
	DBMaxConnLifetime   time.Duration `envconfig:"DB_MAX_CONN_LIFETIME" default:"30m" validate:"required"`
	DBMaxConnIdleTime   time.Duration `envconfig:"DB_MAX_CONN_IDLE_TIME" default:"5m" validate:"required"`
	DBHealthCheckPeriod time.Duration `envconfig:"DB_HEALTH_CHECK_PERIOD" default:"1m" validate:"required"`
	DBPingTimeout       time.Duration `envconfig:"DB_PING_TIMEOUT" default:"5s" validate:"required"`
}

// ServerAddr returns the listening address formatted as ":<port>".
func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

// Load populates configuration from the environment and validates struct constraints.
func Load() (*Config, error) {
	// .env if exists
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// bind to config
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	// validate config
	validate := validator.New("envconfig")
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}
