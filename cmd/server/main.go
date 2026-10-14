package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	grpcadapter "github.com/netologist/go-service-template/internal/adapter/grpc"
	grpchandler "github.com/netologist/go-service-template/internal/adapter/grpc/handler"
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
	return runWithArgs(os.Args[1:])
}

func runWithArgs(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := parseFlags(args, cfg); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	setupLogger(cfg.LogLevel)
	slog.Info("starting application",
		"environment", cfg.Environment,
		"server_mode", cfg.ServerMode,
		"http_enabled", cfg.HTTPEnabled(),
		"grpc_enabled", cfg.GRPCEnabled(),
	)

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

func parseFlags(args []string, cfg *config.Config) error {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	var serverMode string
	fs.StringVar(&serverMode, "server-mode", "", "Execution mode: all, grpc, or rest")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if serverMode != "" {
		switch serverMode {
		case config.ServerModeAll, config.ServerModeGRPC, config.ServerModeREST:
			cfg.ServerMode = serverMode
		default:
			return fmt.Errorf("invalid --server-mode %q: must be one of 'all', 'grpc', or 'rest'", serverMode)
		}
	}
	return nil
}

func runServer(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	healthCheck := handler.NewHealthCheck()
	return runServerWithHealth(ctx, pool, cfg, healthCheck)
}

func runServerWithHealth( //nolint:funlen,cyclop // dual server assembly + graceful shutdown goroutine coordinator
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg *config.Config,
	healthHandler *handler.HealthCheck,
) error {
	itemRepo := postgres.NewItemRepository(pool)
	itemService := usecase.NewItemService(itemRepo)

	var httpServer *http.Server
	if cfg.HTTPEnabled() {
		handlers := httpapi.NewServer(
			healthHandler,
			handler.NewItemHandler(itemService),
		)
		router := httpapi.NewRouterWithOptions(handlers, httpapi.RouterOptions{
			CORSAllowedOrigins: cfg.CORSAllowedOrigins,
			DocsEnabled:        cfg.DocsEnabled(),
		})

		httpServer = &http.Server{
			Addr:              cfg.ServerAddr(),
			Handler:           router,
			ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,
			ReadTimeout:       cfg.ServerReadTimeout,
			WriteTimeout:      cfg.ServerWriteTimeout,
			IdleTimeout:       cfg.ServerIdleTimeout,
		}
	}

	var grpcServer *grpcadapter.Server
	var grpcLis net.Listener
	if cfg.GRPCEnabled() {
		grpcItemHandler := grpchandler.NewItemHandler(itemService)
		grpcServer = grpcadapter.NewServer(grpcItemHandler)

		var lc net.ListenConfig
		lis, err := lc.Listen(ctx, "tcp", cfg.GRPCAddr())
		if err != nil {
			return fmt.Errorf("grpc listen error: %w", err)
		}
		grpcLis = lis
		defer func() { _ = grpcLis.Close() }()
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Goroutine 1: Run HTTP server (if enabled)
	if cfg.HTTPEnabled() && httpServer != nil {
		g.Go(func() error {
			slog.Info("HTTP server listening", "address", httpServer.Addr)
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("server listen error: %w", err)
			}
			return nil
		})
	}

	// Goroutine 2: Run gRPC server (if enabled)
	if cfg.GRPCEnabled() && grpcServer != nil && grpcLis != nil {
		g.Go(func() error {
			slog.Info("gRPC server listening", "address", cfg.GRPCAddr())
			if err := grpcServer.Serve(grpcLis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				return fmt.Errorf("grpc server listen error: %w", err)
			}
			return nil
		})
	}

	// Goroutine 3: Graceful shutdown coordinator
	g.Go(func() error {
		<-gCtx.Done()

		// Context cancellation was triggered by external shutdown signal
		if ctx.Err() != nil {
			slog.Info("shutdown signal received, initiating graceful shutdown")

			// Step 1: Flip readiness probe to unready (K8s stops routing traffic)
			if cfg.HTTPEnabled() && healthHandler != nil {
				healthHandler.SetReady(false)
			}
			if cfg.GRPCEnabled() && grpcServer != nil {
				grpcServer.SetReady(false)
			}

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

		stopGroup, _ := errgroup.WithContext(shutdownCtx)

		if cfg.HTTPEnabled() && httpServer != nil {
			stopGroup.Go(func() error {
				if err := httpServer.Shutdown(shutdownCtx); err != nil {
					return fmt.Errorf("http graceful shutdown failed: %w", err)
				}
				slog.Info("http server shutdown completed successfully")
				return nil
			})
		}

		if cfg.GRPCEnabled() && grpcServer != nil {
			stopGroup.Go(func() error {
				stopped := make(chan struct{})
				go func() {
					grpcServer.GracefulStop()
					close(stopped)
				}()

				select {
				case <-stopped:
					slog.Info("grpc server shutdown completed successfully")
				case <-shutdownCtx.Done():
					slog.Warn("grpc server graceful stop timed out, forcing stop")
					grpcServer.Stop()
				}
				return nil
			})
		}

		if err := stopGroup.Wait(); err != nil {
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
	poolCfg := postgres.PoolConfig{
		DatabaseURL:       cfg.DatabaseURL,
		MaxConns:          cfg.DBMaxConns,
		MinConns:          cfg.DBMinConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheckPeriod,
		PingTimeout:       cfg.DBPingTimeout,
	}

	return postgres.NewPool(ctx, poolCfg)
}
