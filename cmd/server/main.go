package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/netologist/hotel-booking/internal/adapter/http"
	"github.com/netologist/hotel-booking/internal/adapter/http/handler"
)

const (
	DEFAULT_SERVER_PORT        = ":8080"
	DEFAULT_SHUTDOWN_TIMEOUT   = 10 * time.Second
	DEFAULT_READHEADER_TIMEOUT = 30 * time.Second
)

func main() {
	handlers := httpadapter.NewServer(*handler.NewHealthCheck())
	router := httpadapter.NewRouter(handlers)
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
