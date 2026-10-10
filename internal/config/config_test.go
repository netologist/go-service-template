package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/internal/config"
)

const (
	testDatabaseURL          = "postgres://localhost:5432/db"
	errValidationFailed      = "configuration validation failed"
	errProcessEnvFailed      = "failed to process environment variables"
	defaultEnvironment       = "development"
	defaultLogLevel          = "info"
	defaultHTTPPort          = 8080
	defaultServerAddr        = ":8080"
	defaultShutdownTimeout   = 10 * time.Second
	defaultDrainTimeout      = 5 * time.Second
	defaultReadHeaderTimeout = 30 * time.Second
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 2 * time.Minute
	defaultMaxConns          = int32(25)
	defaultMinConns          = int32(2)
	defaultMaxConnLifetime   = 30 * time.Minute
	defaultMaxConnIdleTime   = 5 * time.Minute
	defaultHealthCheckPeriod = 1 * time.Minute
	defaultPingTimeout       = 5 * time.Second

	customEnvironment       = "production"
	customLogLevel          = "warn"
	customHTTPPort          = 9090
	customServerAddr        = ":9090"
	customDatabaseURL       = "postgres://custom:5432/custom_db"
	customShutdownTimeout   = 15 * time.Second
	customDrainTimeout      = 10 * time.Second
	customReadHeaderTimeout = 45 * time.Second
	customReadTimeout       = 20 * time.Second
	customWriteTimeout      = 20 * time.Second
	customIdleTimeout       = 90 * time.Second
	customMaxConns          = int32(50)
	customMinConns          = int32(5)
	customMaxConnLifetime   = 1 * time.Hour
	customMaxConnIdleTime   = 10 * time.Minute
	customHealthCheckPeriod = 2 * time.Minute
	customPingTimeout       = 3 * time.Second
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	if val, ok := os.LookupEnv(key); ok {
		_ = os.Unsetenv(key)
		t.Cleanup(func() {
			_ = os.Setenv(key, val)
		})
	}
}

func resetEnv(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	t.Setenv("DATABASE_URL", testDatabaseURL)
	unsetEnv(t, "APP_ENV")
	unsetEnv(t, "HTTP_PORT")
	unsetEnv(t, "LOG_LEVEL")
	unsetEnv(t, "SERVER_SHUTDOWN_TIMEOUT")
	unsetEnv(t, "SERVER_DRAIN_TIMEOUT")
	unsetEnv(t, "SERVER_READ_HEADER_TIMEOUT")
	unsetEnv(t, "SERVER_READ_TIMEOUT")
	unsetEnv(t, "SERVER_WRITE_TIMEOUT")
	unsetEnv(t, "SERVER_IDLE_TIMEOUT")
	unsetEnv(t, "DB_MAX_CONNS")
	unsetEnv(t, "DB_MIN_CONNS")
	unsetEnv(t, "DB_MAX_CONN_LIFETIME")
	unsetEnv(t, "DB_MAX_CONN_IDLE_TIME")
	unsetEnv(t, "DB_HEALTH_CHECK_PERIOD")
	unsetEnv(t, "DB_PING_TIMEOUT")
}

func TestLoad_Defaults(t *testing.T) {
	resetEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, defaultEnvironment, cfg.Environment)
	assert.Equal(t, defaultHTTPPort, cfg.HTTPPort)
	assert.Equal(t, defaultLogLevel, cfg.LogLevel)
	assert.Equal(t, testDatabaseURL, cfg.DatabaseURL)
	assert.Equal(t, defaultServerAddr, cfg.ServerAddr())
	assert.Equal(t, defaultShutdownTimeout, cfg.ServerShutdownTimeout)
	assert.Equal(t, defaultDrainTimeout, cfg.ServerDrainTimeout)
	assert.Equal(t, defaultReadHeaderTimeout, cfg.ServerReadHeaderTimeout)
	assert.Equal(t, defaultReadTimeout, cfg.ServerReadTimeout)
	assert.Equal(t, defaultWriteTimeout, cfg.ServerWriteTimeout)
	assert.Equal(t, defaultIdleTimeout, cfg.ServerIdleTimeout)
	assert.Equal(t, defaultMaxConns, cfg.DBMaxConns)
	assert.Equal(t, defaultMinConns, cfg.DBMinConns)
	assert.Equal(t, defaultMaxConnLifetime, cfg.DBMaxConnLifetime)
	assert.Equal(t, defaultMaxConnIdleTime, cfg.DBMaxConnIdleTime)
	assert.Equal(t, defaultHealthCheckPeriod, cfg.DBHealthCheckPeriod)
	assert.Equal(t, defaultPingTimeout, cfg.DBPingTimeout)
}

func TestLoad_CustomEnv(t *testing.T) {
	resetEnv(t)
	t.Setenv("APP_ENV", customEnvironment)
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", customLogLevel)
	t.Setenv("DATABASE_URL", customDatabaseURL)
	t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("SERVER_DRAIN_TIMEOUT", "10s")
	t.Setenv("SERVER_READ_HEADER_TIMEOUT", "45s")
	t.Setenv("SERVER_READ_TIMEOUT", "20s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "20s")
	t.Setenv("SERVER_IDLE_TIMEOUT", "90s")
	t.Setenv("DB_MAX_CONNS", "50")
	t.Setenv("DB_MIN_CONNS", "5")
	t.Setenv("DB_MAX_CONN_LIFETIME", "1h")
	t.Setenv("DB_MAX_CONN_IDLE_TIME", "10m")
	t.Setenv("DB_HEALTH_CHECK_PERIOD", "2m")
	t.Setenv("DB_PING_TIMEOUT", "3s")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, customEnvironment, cfg.Environment)
	assert.Equal(t, customHTTPPort, cfg.HTTPPort)
	assert.Equal(t, customLogLevel, cfg.LogLevel)
	assert.Equal(t, customDatabaseURL, cfg.DatabaseURL)
	assert.Equal(t, customServerAddr, cfg.ServerAddr())
	assert.Equal(t, customShutdownTimeout, cfg.ServerShutdownTimeout)
	assert.Equal(t, customDrainTimeout, cfg.ServerDrainTimeout)
	assert.Equal(t, customReadHeaderTimeout, cfg.ServerReadHeaderTimeout)
	assert.Equal(t, customReadTimeout, cfg.ServerReadTimeout)
	assert.Equal(t, customWriteTimeout, cfg.ServerWriteTimeout)
	assert.Equal(t, customIdleTimeout, cfg.ServerIdleTimeout)
	assert.Equal(t, customMaxConns, cfg.DBMaxConns)
	assert.Equal(t, customMinConns, cfg.DBMinConns)
	assert.Equal(t, customMaxConnLifetime, cfg.DBMaxConnLifetime)
	assert.Equal(t, customMaxConnIdleTime, cfg.DBMaxConnIdleTime)
	assert.Equal(t, customHealthCheckPeriod, cfg.DBHealthCheckPeriod)
	assert.Equal(t, customPingTimeout, cfg.DBPingTimeout)
}

type validationCase struct {
	name   string
	setup  func(t *testing.T)
	errSub string
}

func basicValidationCases() []validationCase {
	return []validationCase{
		{
			name: "missing_database_url",
			setup: func(t *testing.T) {
				t.Setenv("DATABASE_URL", "")
			},
			errSub: errValidationFailed,
		},
		{
			name: "invalid_environment",
			setup: func(t *testing.T) {
				t.Setenv("APP_ENV", "invalid_env")
			},
			errSub: errValidationFailed,
		},
		{
			name: "port_too_low",
			setup: func(t *testing.T) {
				t.Setenv("HTTP_PORT", "0")
			},
			errSub: errValidationFailed,
		},
		{
			name: "port_too_high",
			setup: func(t *testing.T) {
				t.Setenv("HTTP_PORT", "65536")
			},
			errSub: errValidationFailed,
		},
		{
			name: "invalid_log_level",
			setup: func(t *testing.T) {
				t.Setenv("LOG_LEVEL", "verbose")
			},
			errSub: errValidationFailed,
		},
	}
}

func dbAndServerValidationCases() []validationCase {
	return []validationCase{
		{
			name: "db_max_conns_zero",
			setup: func(t *testing.T) {
				t.Setenv("DB_MAX_CONNS", "0")
			},
			errSub: errValidationFailed,
		},
		{
			name: "db_min_conns_negative",
			setup: func(t *testing.T) {
				t.Setenv("DB_MIN_CONNS", "-1")
			},
			errSub: errValidationFailed,
		},
		{
			name: "server_shutdown_timeout_zero",
			setup: func(t *testing.T) {
				t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "0s")
			},
			errSub: errValidationFailed,
		},
		{
			name: "server_drain_timeout_negative",
			setup: func(t *testing.T) {
				t.Setenv("SERVER_DRAIN_TIMEOUT", "-1s")
			},
			errSub: errValidationFailed,
		},
		{
			name: "server_read_header_timeout_zero",
			setup: func(t *testing.T) {
				t.Setenv("SERVER_READ_HEADER_TIMEOUT", "0s")
			},
			errSub: errValidationFailed,
		},
		{
			name: "db_max_conn_lifetime_zero",
			setup: func(t *testing.T) {
				t.Setenv("DB_MAX_CONN_LIFETIME", "0s")
			},
			errSub: errValidationFailed,
		},
		{
			name: "db_ping_timeout_zero",
			setup: func(t *testing.T) {
				t.Setenv("DB_PING_TIMEOUT", "0s")
			},
			errSub: errValidationFailed,
		},
	}
}

func TestLoad_ValidationErrors(t *testing.T) {
	testCases := append(basicValidationCases(), dbAndServerValidationCases()...)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			resetEnv(t)
			tt.setup(t)

			cfg, err := config.Load()
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.ErrorContains(t, err, tt.errSub)
		})
	}
}

func TestLoad_InvalidTypes(t *testing.T) {
	t.Run("invalid_port", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("HTTP_PORT", "not-a-number")

		cfg, err := config.Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, errProcessEnvFailed)
	})

	t.Run("invalid_duration", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "invalid-duration")

		cfg, err := config.Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, errProcessEnvFailed)
	})
}

func TestLoad_DrainTimeoutZeroAllowed(t *testing.T) {
	resetEnv(t)
	t.Setenv("SERVER_DRAIN_TIMEOUT", "0s")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, time.Duration(0), cfg.ServerDrainTimeout)
}

func TestLoad_CORSAllowedOrigins(t *testing.T) {
	t.Run("valid origins", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,http://localhost:3000")

		cfg, err := config.Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, []string{"https://app.example.com", "http://localhost:3000"}, cfg.CORSAllowedOrigins)
	})

	t.Run("wildcard rejected", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://*,http://localhost:3000")

		cfg, err := config.Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, errValidationFailed)
	})

	t.Run("invalid url rejected", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("CORS_ALLOWED_ORIGINS", "not-a-valid-url")

		cfg, err := config.Load()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, errValidationFailed)
	})
}

func TestConfig_DocsEnabled(t *testing.T) {
	cfg := &config.Config{Environment: "development"}
	assert.True(t, cfg.DocsEnabled())

	cfg.Environment = "staging"
	assert.True(t, cfg.DocsEnabled())

	cfg.Environment = "test"
	assert.True(t, cfg.DocsEnabled())

	cfg.Environment = "production"
	assert.False(t, cfg.DocsEnabled())
}
