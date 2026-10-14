package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	environmentProduction = "production"

	ServerModeAll  = "all"
	ServerModeGRPC = "grpc"
	ServerModeREST = "rest"
)

type Config struct {
	Environment string `env:"APP_ENV" envDefault:"development" validate:"required,oneof=development staging production test"` //nolint:lll // struct tags cannot be split across lines
	ServerMode  string `env:"SERVER_MODE" envDefault:"all" validate:"required,oneof=all grpc rest"`
	HTTPPort    int    `env:"HTTP_PORT" envDefault:"8080" validate:"required,min=1,max=65535"`
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"50051" validate:"required,min=1,max=65535"`
	DatabaseURL string `env:"DATABASE_URL" validate:"required"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info" validate:"required,oneof=debug info warn error"`

	// CORSAllowedOrigins lists the exact origins (scheme://host[:port]) allowed to call the API from a browser.
	// Empty disables CORS entirely. Wildcards are rejected.
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," validate:"dive,http_url,excludes=*"` //nolint:lll // struct tags cannot be split across lines

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

// GRPCAddr returns the gRPC listening address formatted as ":<port>".
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%d", c.GRPCPort)
}

// HTTPEnabled reports whether the HTTP/REST server should run.
func (c *Config) HTTPEnabled() bool {
	return c.ServerMode == ServerModeAll || c.ServerMode == ServerModeREST
}

// GRPCEnabled reports whether the gRPC server should run.
func (c *Config) GRPCEnabled() bool {
	return c.ServerMode == ServerModeAll || c.ServerMode == ServerModeGRPC
}

// DocsEnabled reports whether the Swagger UI and the OpenAPI document are served.
// They are never exposed in production.
func (c *Config) DocsEnabled() bool {
	return c.Environment != environmentProduction
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
