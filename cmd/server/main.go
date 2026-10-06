package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/netologist/hotel-booking/internal/adapter/rest"
	"github.com/netologist/hotel-booking/internal/adapter/rest/handler"
)

const (
	DEFAULT_SERVER_PORT        = ":8080"
	DEFAULT_SHUTDOWN_TIMEOUT   = 10 * time.Second
	DEFAULT_READHEADER_TIMEOUT = 30 * time.Second
)

func main() {
	handlers := rest.NewServer(*handler.NewHealthCheck())
	router := rest.NewRouter(handlers)
	server := &http.Server{
		Addr:              DEFAULT_SERVER_PORT,
		ReadHeaderTimeout: DEFAULT_READHEADER_TIMEOUT,
		Handler:           router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		DEFAULT_SHUTDOWN_TIMEOUT,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
