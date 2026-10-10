package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/netologist/go-service-template/internal/adapter/httpapi"
	"github.com/netologist/go-service-template/internal/adapter/httpapi/handler"
	"github.com/netologist/go-service-template/internal/adapter/postgres"
	"github.com/netologist/go-service-template/internal/config"
	"github.com/netologist/go-service-template/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application terminated with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	setupLogger(cfg.LogLevel)
	slog.Info("starting application", "environment", cfg.Environment)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := setupPostgresPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	slog.Info("connected to PostgreSQL database successfully")

	if err := runServer(ctx, pool, cfg); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func runServer(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	healthCheck := handler.NewHealthCheck()
	return runServerWithHealth(ctx, pool, cfg, healthCheck)
}

func runServerWithHealth( //nolint:funlen // server assembly + graceful shutdown goroutine coordinator
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg *config.Config,
	healthHandler *handler.HealthCheck,
) error {
	hotelRepo := postgres.NewHotelRepository(pool)
	hotelService := usecase.NewHotelService(hotelRepo)
	handlers := httpapi.NewServer(
		healthHandler,
		handler.NewHotelHandler(hotelService),
	)
	router := httpapi.NewRouterWithOptions(handlers, httpapi.RouterOptions{
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		DocsEnabled:        cfg.DocsEnabled(),
	})

	server := &http.Server{
		Addr:              cfg.ServerAddr(),
		Handler:           router,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		ReadTimeout:       cfg.ServerReadTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
		IdleTimeout:       cfg.ServerIdleTimeout,
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Goroutine 1: Run HTTP server
	g.Go(func() error {
		slog.Info("HTTP server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server listen error: %w", err)
		}
		return nil
	})

	// Goroutine 2: Graceful shutdown coordinator
	g.Go(func() error {
		<-gCtx.Done()

		// Context cancellation was triggered by external shutdown signal
		if ctx.Err() != nil {
			slog.Info("shutdown signal received, initiating graceful shutdown")

			// Step 1: Flip readiness probe to unready (K8s stops routing traffic)
			healthHandler.SetReady(false)

			// Step 2: Wait for connection drain (K8s endpoint propagation window)
			if cfg.ServerDrainTimeout > 0 {
				slog.Info("waiting for connection drain before shutdown", "duration", cfg.ServerDrainTimeout)
				drainTimer := time.NewTimer(cfg.ServerDrainTimeout)
				defer drainTimer.Stop()
				<-drainTimer.C
			}
		}

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ServerShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		slog.Info("server shutdown completed successfully")
		return nil
	})

	return g.Wait()
}

func setupLogger(levelStr string) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		level = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

func setupPostgresPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return postgres.NewPool(ctx, postgres.PoolConfig{
		DatabaseURL:       cfg.DatabaseURL,
		MaxConns:          cfg.DBMaxConns,
		MinConns:          cfg.DBMinConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheckPeriod,
		PingTimeout:       cfg.DBPingTimeout,
	})
}
