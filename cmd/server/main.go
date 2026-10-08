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

	"github.com/jackc/pgx/v5/pgxpool"
	httpadapter "github.com/netologist/hotel-booking/internal/adapter/http"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
	"github.com/netologist/hotel-booking/internal/adapter/postgres"
	"github.com/netologist/hotel-booking/internal/adapter/validator"
	"github.com/netologist/hotel-booking/internal/config"
	"github.com/netologist/hotel-booking/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := setupPostgresPool(ctx, cfg)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		return
	}
	defer pool.Close()
	slog.Info("connected to PostgreSQL database successfully")

	if err := runServer(ctx, pool, cfg); err != nil {
		slog.Error("server error", "error", err)
	}
}

func runServer(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	appValidator := validator.NewForJsonTag()
	hotelRepo := postgres.NewHotelRepository(pool)
	hotelService := usecase.NewHotelService(hotelRepo)
	handlers := httpadapter.NewServer(
		handler.NewHealthCheck(),
		handler.NewHotelHandler(hotelService, appValidator),
	)
	router := httpadapter.NewRouter(handlers)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
		ReadTimeout:       cfg.ServerReadTimeout,
		WriteTimeout:      cfg.ServerWriteTimeout,
		IdleTimeout:       cfg.ServerIdleTimeout,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	select {
	case err := <-serverErrCh:
		return fmt.Errorf("server listen error: %w", err)
	case <-ctx.Done():
		slog.Info("shutdown signal received, initiating graceful shutdown")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ServerShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("server shutdown completed successfully")
	return nil
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
	return postgres.NewPool(ctx,
		postgres.WithDatabaseURL(cfg.DatabaseURL),
		postgres.WithMaxConns(cfg.DBMaxConns),
		postgres.WithMinConns(cfg.DBMinConns),
		postgres.WithMaxConnLifetime(cfg.DBMaxConnLifetime),
		postgres.WithMaxConnIdleTime(cfg.DBMaxConnIdleTime),
		postgres.WithHealthCheckPeriod(cfg.DBHealthCheckPeriod),
		postgres.WithPingTimeout(cfg.DBPingTimeout),
	)
}
