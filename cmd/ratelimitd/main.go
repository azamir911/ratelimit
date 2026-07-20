package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azamir911/ratelimit"
	"github.com/azamir911/ratelimit/internal/httpapi"
)

func main() {
	var (
		address         = flag.String("listen", ":8080", "HTTP listen address")
		limit           = flag.Uint64("limit", 100, "maximum accumulated cost per key and window")
		window          = flag.Duration("window", time.Minute, "fixed-window duration")
		maxKeys         = flag.Uint64("max-keys", 100_000, "maximum retained keys")
		cleanupInterval = flag.Duration("cleanup-interval", time.Minute, "expired-key cleanup interval")
	)
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	limiter, err := ratelimit.New(ratelimit.Config{
		Limit:           *limit,
		Window:          *window,
		MaxKeys:         *maxKeys,
		CleanupInterval: *cleanupInterval,
	})
	if err != nil {
		logger.Error("invalid limiter configuration", "error", err)
		os.Exit(2)
	}
	defer limiter.Close()

	server := &http.Server{
		Addr:              *address,
		Handler:           httpapi.NewHandler(limiter),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("rate-limit service started", "address", *address)
		serverErrors <- server.ListenAndServe()
	}()

	signalContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalContext.Done():
		logger.Info("shutdown requested")
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("shutdown completed")
}
